package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCommitCoordinator_Ordered 顺序完成的场景：逐条提交，无跳号。
func TestCommitCoordinator_Ordered(t *testing.T) {
	c := newCommitCoordinator()
	c.init(0, 0)
	for i := int64(0); i < 5; i++ {
		to, ok := c.markDone(0, i)
		if !ok {
			t.Fatalf("顺序完成时 offset=%d 应可提交", i)
		}
		if to != i {
			t.Fatalf("期望提交 offset=%d, 实际 %d", i, to)
		}
	}
}

// TestCommitCoordinator_OutOfOrder 乱序完成的场景：只提交连续前缀，等待补齐。
func TestCommitCoordinator_OutOfOrder(t *testing.T) {
	c := newCommitCoordinator()
	c.init(0, 0)

	// offset 2,3 先完成，必须等待
	if _, ok := c.markDone(0, 2); ok {
		t.Fatal("offset=2 前序未完成，不应提交")
	}
	if _, ok := c.markDone(0, 3); ok {
		t.Fatal("offset=3 前序未完成，不应提交")
	}
	// offset 1 完成，仍缺 offset 0
	if _, ok := c.markDone(0, 1); ok {
		t.Fatal("offset=0 未完成，不应提交")
	}
	// offset 0 完成 → 连续前缀 0,1,2,3 全部补齐，提交到 3
	to, ok := c.markDone(0, 0)
	if !ok {
		t.Fatal("offset=0 完成后应触发提交")
	}
	if to != 3 {
		t.Fatalf("期望提交 offset=3, 实际 %d", to)
	}
	// offset 4 完成 → 提交 4
	to, ok = c.markDone(0, 4)
	if !ok || to != 4 {
		t.Fatalf("期望提交 offset=4, 实际 to=%d ok=%v", to, ok)
	}
}

// TestCommitCoordinator_MultiPartition 多 partition 状态互相隔离。
func TestCommitCoordinator_MultiPartition(t *testing.T) {
	c := newCommitCoordinator()
	c.init(0, 0)
	c.init(1, 100)

	// partition=0 与 partition=1 各自的 base 独立
	if to, ok := c.markDone(0, 1); ok || to != 0 {
		t.Fatalf("partition=0 的 offset=1 不应提交, to=%d ok=%v", to, ok)
	}
	if to, ok := c.markDone(1, 101); ok || to != 0 {
		t.Fatalf("partition=1 的 offset=101 不应提交, to=%d ok=%v", to, ok)
	}
	// partition=1 补齐 100 后，连续前缀 100,101 一并提交到 101
	if to, ok := c.markDone(1, 100); !ok || to != 101 {
		t.Fatalf("partition=1 应提交到 101, to=%d ok=%v", to, ok)
	}
	// partition=0 补齐 0 后，连续前缀 0,1 一并提交到 1
	if to, ok := c.markDone(0, 0); !ok || to != 1 {
		t.Fatalf("partition=0 应提交到 1, to=%d ok=%v", to, ok)
	}
}

// TestCommitCoordinator_NonZeroStart 起点 offset 非 0（已提交过部分消息的场景）。
func TestCommitCoordinator_NonZeroStart(t *testing.T) {
	c := newCommitCoordinator()
	c.init(0, 100) // 之前已提交到 99

	// offset 100 先完成 → 可提交
	if to, ok := c.markDone(0, 100); !ok || to != 100 {
		t.Fatalf("期望提交 offset=100, to=%d ok=%v", to, ok)
	}
	// 旧 offset（99 及以下）应被忽略
	if _, ok := c.markDone(0, 99); ok {
		t.Fatal("已提交过的旧 offset 不应再次提交")
	}
}

// mockReader 测试用消息读取器：按顺序吐出事件，取完阻塞在 ctx 上。
type mockReader struct {
	mu        sync.Mutex
	events    []*Event
	committed []*Message
}

func (m *mockReader) add(e *Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
}

func (m *mockReader) Fetch(ctx context.Context) (*Event, error) {
	m.mu.Lock()
	if len(m.events) > 0 {
		e := m.events[0]
		m.events = m.events[1:]
		m.mu.Unlock()
		return e, nil
	}
	m.mu.Unlock()
	<-ctx.Done()
	return nil, ctx.Err()
}

func (m *mockReader) Commit(ctx context.Context, msg *Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.committed = append(m.committed, msg)
	return nil
}

func (m *mockReader) Close() error { return nil }

// waitFor 轮询等待条件满足。
func waitFor(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

// TestPartition_ConcurrentCommitNoLoss 集成测试：消息并发处理、完成顺序乱序，
// 但 commit 必须始终只推进"连续前缀"，最终全部处理后提交到最大 offset，无跳号。
func TestPartition_ConcurrentCommitNoLoss(t *testing.T) {
	const total = 30
	reader := &mockReader{}
	var processed int32
	handler := ConsumerHandlerFunc(func(ctx context.Context, e *Event) error {
		// 按 key 首字节制造不同处理耗时，让完成顺序乱序
		time.Sleep(time.Duration(e.Msg.Key[0]%4) * time.Millisecond)
		atomic.AddInt32(&processed, 1)
		return nil
	})

	pool := NewWorkerPool(4, 64, handler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)
	part := NewPartition(reader, pool)
	part.Start(ctx)

	for i := 0; i < total; i++ {
		reader.add(&Event{Msg: &Message{
			Topic:     "topic",
			Partition: 0,
			Offset:    int64(i),
			Key:       []byte(fmt.Sprintf("key-%d", i)),
			Value:     []byte(fmt.Sprintf("%d", i)),
		}})
	}

	if !waitFor(5*time.Second, func() bool { return atomic.LoadInt32(&processed) == total }) {
		t.Fatalf("消息未全部处理: processed=%d", atomic.LoadInt32(&processed))
	}

	part.Stop()
	pool.StopAndWait()

	reader.mu.Lock()
	defer reader.mu.Unlock()
	if len(reader.committed) == 0 {
		t.Fatal("没有任何提交")
	}
	last := reader.committed[len(reader.committed)-1].Offset
	if last != total-1 {
		t.Fatalf("最终应提交到 offset=%d, 实际 %d", total-1, last)
	}
	// commit 序列必须严格递增（不跳号、不回退）：每个 commit 都是连续前缀的终点
	for i := 1; i < len(reader.committed); i++ {
		if reader.committed[i].Offset <= reader.committed[i-1].Offset {
			t.Fatalf("commit 序列异常: %d -> %d", reader.committed[i-1].Offset, reader.committed[i].Offset)
		}
	}
}

// TestPartition_OneMessageFails_NoLoss 集成测试：某条消息永久处理失败，
// 提交必须停在该消息之前（不提交更后面的 offset），保证崩溃后失败消息会重新投递。
func TestPartition_OneMessageFails_NoLoss(t *testing.T) {
	const total = 20
	const failOffset = int64(7)
	reader := &mockReader{}
	var processed int32
	handler := ConsumerHandlerFunc(func(ctx context.Context, e *Event) error {
		atomic.AddInt32(&processed, 1)
		if e.Msg.Offset == failOffset {
			return errors.New("simulated permanent failure")
		}
		return nil
	})

	pool := NewWorkerPool(4, 64, handler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)
	part := NewPartition(reader, pool)
	part.Start(ctx)

	for i := 0; i < total; i++ {
		reader.add(&Event{Msg: &Message{
			Topic:     "topic",
			Partition: 0,
			Offset:    int64(i),
			Key:       []byte(fmt.Sprintf("key-%d", i)),
			Value:     []byte(fmt.Sprintf("%d", i)),
		}})
	}

	// 等全部处理完（失败的也算已尝试）
	if !waitFor(5*time.Second, func() bool { return atomic.LoadInt32(&processed) == total }) {
		t.Fatalf("消息未全部处理: processed=%d", atomic.LoadInt32(&processed))
	}

	part.Stop()
	pool.StopAndWait()

	reader.mu.Lock()
	defer reader.mu.Unlock()
	if len(reader.committed) == 0 {
		t.Fatal("没有任何提交")
	}
	last := reader.committed[len(reader.committed)-1].Offset
	if last >= failOffset {
		t.Fatalf("offset=%d 处理失败, 提交绝不能越过它: 实际最后提交=%d", failOffset, last)
	}
	// 并且 0..failOffset-1 必须都已提交（连续前缀）
	covered := make([]bool, total)
	for _, c := range reader.committed {
		for o := int64(0); o <= c.Offset; o++ {
			covered[o] = true
		}
	}
	for i := int64(0); i < failOffset; i++ {
		if !covered[i] {
			t.Fatalf("offset=%d 已成功处理但未提交（丢消息）", i)
		}
	}
}

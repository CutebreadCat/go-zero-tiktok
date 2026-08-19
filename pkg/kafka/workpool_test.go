package kafka

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestWorkerPool_SameKeyOrdered 验证：同一 key 的消息由同一 worker 串行处理、顺序严格保持。
//
// 模拟真实业务：key = target_id（固定不变），序号放在 Value 中区分事件先后。
// 验证思路：
// 1) 保序：每个 key 消费的 Value 序号必须严格递增（同一 worker 串行消费同一分片 FIFO）；
// 2) 无并发：对每个 key 加锁，若同一 key 被并发消费，TryLock 会失败暴露问题。
func TestWorkerPool_SameKeyOrdered(t *testing.T) {
	const workers = 4
	const keyCount = 8
	const rounds = 50

	keyLocks := make([]sync.Mutex, keyCount)

	// 记录消费是否出现乱序：每个 key 必须按 Value 序号(0..rounds-1)被消费
	var mu sync.Mutex
	nextSeq := make([]int, keyCount) // 每个 key 期望的下一个序号
	concurrentHit := int32(0)        // 检测到并发消费同一 key 的次数

	handler := ConsumerHandlerFunc(func(ctx context.Context, e *Event) error {
		var ki, seq int
		fmt.Sscanf(string(e.Msg.Key), "target-%d", &ki)
		fmt.Sscanf(string(e.Msg.Value), "%d", &seq)

		// 检测并发：非阻塞抢锁，抢不到说明同一 key 正被另一个 worker 消费
		if !keyLocks[ki].TryLock() {
			atomic.AddInt32(&concurrentHit, 1)
			return nil
		}
		defer keyLocks[ki].Unlock()

		mu.Lock()
		defer mu.Unlock()
		if nextSeq[ki] != seq {
			t.Errorf("key=%d 乱序: 期望序号 %d, 实际 %d", ki, nextSeq[ki], seq)
		}
		nextSeq[ki] = seq + 1
		return nil
	})

	pool := NewWorkerPool(workers, 64, handler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	// 串行提交：每轮把 8 个 target 各提交一次，key 固定，序号在 Value 里
	for r := 0; r < rounds; r++ {
		for ki := 0; ki < keyCount; ki++ {
			pool.Submit(&Event{
				Msg: &Message{
					Key:   []byte(fmt.Sprintf("target-%d", ki)),
					Value: []byte(fmt.Sprintf("%d", r)),
				},
			}, nil)
		}
	}

	// 等待全部消费完：轮询 nextSeq 是否全部达到 rounds
	deadline := time.After(3 * time.Second)
	done := make(chan struct{})
	go func() {
		for {
			mu.Lock()
			allDone := true
			for ki := 0; ki < keyCount; ki++ {
				if nextSeq[ki] != rounds {
					allDone = false
					break
				}
			}
			mu.Unlock()
			if allDone {
				close(done)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-done:
	case <-deadline:
		mu.Lock()
		defer mu.Unlock()
		t.Fatalf("消费未完成, nextSeq=%v", nextSeq)
	}

	if atomic.LoadInt32(&concurrentHit) > 0 {
		t.Errorf("同一 key 被并发消费 %d 次, 分片保序失效", concurrentHit)
	}
}

// TestWorkerPool_SameKeySameShard 验证 shardIndex：同一 key 永远路由到同一分片。
func TestWorkerPool_SameKeySameShard(t *testing.T) {
	pool := NewWorkerPool(8, 64, nil)
	seen := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("target-%d", i%50)
		idx := pool.shardIndex([]byte(key))
		if prev, ok := seen[key]; ok {
			if prev != idx {
				t.Fatalf("key %s 路由漂移: %d -> %d", key, prev, idx)
			}
		} else {
			seen[key] = idx
		}
	}
}

// TestWorkerPool_NilKeyRoundRobin 验证无 key 消息轮询分散到所有分片，不挤到分片 0。
func TestWorkerPool_NilKeyRoundRobin(t *testing.T) {
	pool := NewWorkerPool(8, 64, nil)
	seen := make(map[int]bool)
	const n = 32
	for i := 0; i < n; i++ {
		seen[pool.shardIndex(nil)] = true
	}
	if len(seen) != 8 {
		t.Fatalf("期望轮询到全部 8 个分片, 实际 %d: %v", len(seen), seen)
	}
}

// TestWorkerPool_ShardDistribution 验证多 key 在分片间分布相对均匀（无明显倾斜）。
func TestWorkerPool_ShardDistribution(t *testing.T) {
	pool := NewWorkerPool(4, 64, nil)
	counts := make([]int, 4)
	for i := 0; i < 4000; i++ {
		counts[pool.shardIndex([]byte(fmt.Sprintf("target-%d", i)))]++
	}
	sort.Ints(counts)
	avg := 4000 / 4
	if counts[3] > avg*2 {
		t.Fatalf("分片分布倾斜: %v (avg=%d)", counts, avg)
	}
}

// TestWorkerPool_QueueFullBackpressure 验证队列满时 Submit 阻塞（背压），不丢消息。
func TestWorkerPool_QueueFullBackpressure(t *testing.T) {
	var consumed int32
	handler := ConsumerHandlerFunc(func(ctx context.Context, e *Event) error {
		atomic.AddInt32(&consumed, 1)
		time.Sleep(5 * time.Millisecond)
		return nil
	})
	pool := NewWorkerPool(1, 2, handler) // 1 worker, 每分片缓冲 2
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	// 大量提交：队列满会阻塞，但全部必须被消费
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			pool.Submit(&Event{Msg: &Message{Key: []byte("k")}}, nil)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Submit 卡死, 背压失效")
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&consumed) == 500 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("消费不完整: consumed=%d", atomic.LoadInt32(&consumed))
}

// TestWorkerPool_StopAndWait_Drain 验证优雅关闭：StopAndWait 后已提交任务全部处理完，
// 且不再接受新提交（返回 ErrPoolClosed）。
func TestWorkerPool_StopAndWait_Drain(t *testing.T) {
	const total = 100
	var consumed int32
	handler := ConsumerHandlerFunc(func(ctx context.Context, e *Event) error {
		time.Sleep(2 * time.Millisecond)
		atomic.AddInt32(&consumed, 1)
		return nil
	})
	pool := NewWorkerPool(2, 64, handler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	for i := 0; i < total; i++ {
		pool.Submit(&Event{Msg: &Message{Key: []byte("k")}}, nil)
	}

	pool.StopAndWait()

	if atomic.LoadInt32(&consumed) != total {
		t.Fatalf("StopAndWait 后应处理完全部 %d 条, 实际 %d", total, atomic.LoadInt32(&consumed))
	}
	if err := pool.SubmitCtx(context.Background(), &Event{Msg: &Message{Key: []byte("k")}}, nil); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("池已停止后提交应返回 ErrPoolClosed, 实际 %v", err)
	}
}

// TestWorkerPool_PanicSelfHeal 验证 handler panic 不会杀死 worker：
// panic 被 recover 兜底后，后续任务仍被正常消费。
func TestWorkerPool_PanicSelfHeal(t *testing.T) {
	const total = 100
	var invoked int32
	var commits int32
	handler := ConsumerHandlerFunc(func(ctx context.Context, e *Event) error {
		n := atomic.AddInt32(&invoked, 1)
		if n <= 3 {
			panic("simulated handler panic")
		}
		return nil
	})
	pool := NewWorkerPool(2, 64, handler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	for i := 0; i < total; i++ {
		pool.Submit(&Event{Msg: &Message{Key: []byte("k")}}, func() { atomic.AddInt32(&commits, 1) })
	}

	if !waitFor(5*time.Second, func() bool { return atomic.LoadInt32(&invoked) == total }) {
		t.Fatalf("消息未全部被 handler 调用: invoked=%d", atomic.LoadInt32(&invoked))
	}
	// panic 的 3 条不应提交（at-least-once 重投），其余 97 条应提交
	if atomic.LoadInt32(&commits) != total-3 {
		t.Fatalf("期望提交 %d 条, 实际 %d", total-3, atomic.LoadInt32(&commits))
	}
	pool.StopAndWait()
}

// TestWorkerPool_SubmitCtxCancel 验证 SubmitCtx 响应 ctx 取消：
// 队列满阻塞时 ctx 取消应立即返回，且消息未入队。
func TestWorkerPool_SubmitCtxCancel(t *testing.T) {
	handler := ConsumerHandlerFunc(func(ctx context.Context, e *Event) error {
		time.Sleep(100 * time.Millisecond) // 让 worker 保持忙碌
		return nil
	})
	pool := NewWorkerPool(1, 1, handler) // 1 worker + 缓冲 1 → 提交第 3 条会阻塞
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	// 先占满 worker + 缓冲
	pool.Submit(&Event{Msg: &Message{Key: []byte("a")}}, nil)
	pool.Submit(&Event{Msg: &Message{Key: []byte("b")}}, nil)

	submitCtx, submitCancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		submitCancel()
	}()

	start := time.Now()
	err := pool.SubmitCtx(submitCtx, &Event{Msg: &Message{Key: []byte("c")}}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("期望返回 context.Canceled, 实际 %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("SubmitCtx 未及时响应取消")
	}
	pool.StopAndWait()
}

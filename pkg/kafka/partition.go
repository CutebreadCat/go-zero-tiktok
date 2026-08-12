package kafka

import (
	"context"
	"errors"
	"sync"
	"time"

	appLogger "go_zero-tiktok/pkg/logger"
)

// MessageReader 消息读取器接口，由 KafkaReader 实现
type MessageReader interface {
	Fetch(context.Context) (*Event, error)
	Commit(context.Context, *Message) error
}

// CloseableReader 可关闭的读取器（用于优雅关闭时解除 Fetch 阻塞）
type CloseableReader interface {
	Close() error
}

// commitCoordinator 保证同一 partition 的 offset 按"连续前缀"提交。
//
// 背景：分片并发处理会让同一 partition 各条消息的处理完成顺序乱序。若逐条直接提交，
// 高 offset 可能先于低 offset 提交，进程崩溃后低 offset 消息被 broker 认为已消费而被跳过（丢失）。
// 该协调器只提交"已成功处理的最长连续前缀"，保证 at-least-once 语义下不丢消息。
//
// 使用约定：fetcher 每 fetch 到一条消息必须先调用 init（按 offset 升序调用），
// 再提交给 pool；worker 处理成功后调用 markDone。init 先于同 partition 的任意 markDone。
type commitCoordinator struct {
	mu          sync.Mutex
	pending     map[int]map[int64]struct{} // partition -> 已处理但未达到提交条件的 offset
	base        map[int]int64              // partition -> 下一个应提交的 offset
	initialized map[int]bool               // partition -> base 是否已初始化
}

func newCommitCoordinator() *commitCoordinator {
	return &commitCoordinator{
		pending:     make(map[int]map[int64]struct{}),
		base:        make(map[int]int64),
		initialized: make(map[int]bool),
	}
}

// init 初始化 partition 的起点 offset（首个被 fetch 到的 offset）。
// fetch 按 partition 内 offset 升序，因此首个 init 值即该 partition 的消费起点；
// 之后重复调用为 no-op。
func (c *commitCoordinator) init(partition int, offset int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.initialized[partition] {
		return
	}
	c.initialized[partition] = true
	c.base[partition] = offset
}

// markDone 记录 partition 的 offset 已处理成功。
// 若连续前缀推进，返回最高可提交的 offset 与 true；否则返回 false（等待前序补上）。
func (c *commitCoordinator) markDone(partition int, offset int64) (int64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized[partition] {
		// 防御：理论上不会发生（fetcher 先 init 后 submit）。不提交，避免跳过前序。
		c.initialized[partition] = true
		c.base[partition] = offset
		return 0, false
	}

	base := c.base[partition]
	if offset < base {
		// 已提交过，忽略（rebalance 重投等场景）
		return 0, false
	}
	if offset != base {
		// 前序未处理完，先记入 pending 等待补齐
		if c.pending[partition] == nil {
			c.pending[partition] = make(map[int64]struct{})
		}
		c.pending[partition][offset] = struct{}{}
		return 0, false
	}

	// offset == base：连续，向前推进到第一个空洞
	next := offset + 1
	for {
		if _, ok := c.pending[partition][next]; ok {
			delete(c.pending[partition], next)
			next++
			continue
		}
		break
	}
	c.base[partition] = next
	return next - 1, true
}

type Partition struct {
	reader MessageReader
	pool   *WorkerPool

	stopOnce sync.Once
	stopCh   chan struct{}
}

func NewPartition(reader MessageReader, pool *WorkerPool) *Partition {
	return &Partition{reader: reader, pool: pool, stopCh: make(chan struct{})}
}

func (p *Partition) Start(ctx context.Context) {
	appLogger.Info("partition fetcher starting")
	go func() {
		coord := newCommitCoordinator()
		appLogger.Info("partition fetcher goroutine started")
		for {
			select {
			case <-ctx.Done():
				appLogger.Infof("partition fetcher stopped: %v", ctx.Err())
				return
			case <-p.stopCh:
				appLogger.Info("partition fetcher stopped by Stop")
				return
			default:
			}

			event, err := p.reader.Fetch(ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				appLogger.Errorf("partition fetch failed: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-p.stopCh:
					return
				case <-time.After(partitionFetchRetryWait):
				}
				continue
			}
			if event == nil || event.Msg == nil {
				continue
			}

			msg := event.Msg
			// 必须先于 submit：按 fetch 的 offset 升序初始化 partition 消费起点，
			// 保证并发处理下 commit 只推进"已处理的最长连续前缀"。
			coord.init(msg.Partition, msg.Offset)
			if err := p.pool.SubmitCtx(ctx, event, p.commitFunc(ctx, coord, msg)); err != nil {
				appLogger.Infof("partition fetcher stopped submitting: %v", err)
				return
			}
		}
	}()
}

// commitFunc 生成该消息的提交回调：经 commitCoordinator 协调，只提交"连续前缀"。
// 同一 partition 前序消息未处理完时本次不提交，等待前序补齐后统一提交最高前缀。
func (p *Partition) commitFunc(ctx context.Context, coord *commitCoordinator, msg *Message) CommitFunc {
	return func() {
		if msg == nil || msg.Partition < 0 {
			return
		}
		to, ok := coord.markDone(msg.Partition, msg.Offset)
		if !ok {
			return
		}
		if err := p.reader.Commit(ctx, &Message{Topic: msg.Topic, Partition: msg.Partition, Offset: to}); err != nil {
			appLogger.Errorf("commit message failed: partition=%d offset=%d err=%v", msg.Partition, to, err)
		} else {
			appLogger.Infof("message committed: partition=%d offset=%d", msg.Partition, to)
		}
	}
}

// Stop 停止 fetcher：不再从 broker 拉取新消息、不再向 pool 提交。
func (p *Partition) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

// Close 关闭底层 reader（可解除阻塞中的 Fetch）。
func (p *Partition) Close() error {
	if c, ok := p.reader.(CloseableReader); ok {
		return c.Close()
	}
	return nil
}

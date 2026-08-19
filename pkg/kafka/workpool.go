package kafka

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"runtime/debug"
	"sync"
	"sync/atomic"

	appLogger "go_zero-tiktok/pkg/logger"
)

// CommitFunc 提交 Offset 的回调函数类型
// 由 Partition 层封装，WorkerPool 在消息处理成功后调用
type CommitFunc func()

// Job WorkerPool 的任务单元，包含业务消息和提交回调
type Job struct {
	Msg    *Event
	Commit CommitFunc
}

// ErrPoolClosed 提交到已停止的 WorkerPool 时返回
var ErrPoolClosed = errors.New("kafka: worker pool is stopped")

// WorkerPool 分片 Worker 池。
//
// 设计要点：按消息 Key 哈希路由到固定分片队列，每个 Worker 绑定一个分片，
// 实现"同 key 必进同分片、由同一 Worker 串行处理"，从而保证单 target 有序；
// 不同 key 分散到不同分片，由不同 Worker 并行处理，最大化吞吐。
// 这是"顺序保证"与"消费并行度"之间业界通用解法（分片/分区内串行、分区间并行）。
type WorkerPool struct {
	shards  []chan Job // 分片队列，len(shards) == worker 数
	next    uint64     // 无 key 消息的轮询计数器
	handler ConsumerHandler

	wg       sync.WaitGroup
	stopOnce sync.Once
	stopCh   chan struct{} // StopAndWait 关闭，worker 收到后进入 drain 模式
}

// NewWorkerPool 创建分片 Worker 池。
// workers: worker 数 = 分片数（一个 worker 独占一个分片队列）；
// queueSize: 每个分片队列的缓冲大小（注意：总缓冲 = workers * queueSize）；
// h: 业务消费处理器。
func NewWorkerPool(workers int, queueSize int, h ConsumerHandler) *WorkerPool {
	if workers <= 0 {
		workers = DefaultConsumerWorkerCount
	}
	if queueSize <= 0 {
		queueSize = DefaultConsumerQueueSize
	}
	shards := make([]chan Job, workers)
	for i := range shards {
		shards[i] = make(chan Job, queueSize)
	}
	return &WorkerPool{
		shards:  shards,
		handler: h,
		stopCh:  make(chan struct{}),
	}
}

// shardIndex 计算消息 Key 对应的分片下标。
// 有 key：FNV-1a 哈希取模，同 key 必同分片；
// 无 key：轮询（原子计数器），避免全部挤到分片 0。
func (p *WorkerPool) shardIndex(key []byte) int {
	n := uint64(len(p.shards))
	if len(key) == 0 {
		seq := atomic.AddUint64(&p.next, 1)
		return int((seq - 1) % n)
	}
	h := fnv.New32a()
	h.Write(key)
	return int(uint64(h.Sum32()) % n)
}

// Start 启动 workers。每个 Worker 固定消费自己的分片队列。
// ctx 取消或调用 StopAndWait 后，Worker 进入 drain 模式：把已提交的任务处理完再退出。
func (p *WorkerPool) Start(ctx context.Context) {
	for i := range p.shards {
		shard := p.shards[i]
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			p.runWorker(ctx, workerID, shard)
		}(i)
	}
}

// runWorker worker 主循环：正常模式下阻塞等待任务；
// 收到 ctx 取消或 stopCh 关闭信号后切换为 drain 模式，消费完队列中剩余任务再退出。
// handler panic 时由 consumeJob 兜底转 error，worker 循环不会死亡（自愈）。
func (p *WorkerPool) runWorker(ctx context.Context, workerID int, shard <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			p.drain(workerID, shard)
			return
		case <-p.stopCh:
			p.drain(workerID, shard)
			return
		case job := <-shard:
			p.consumeJob(job)
		}
	}
}

// drain 停止接收新任务后，把分片队列中已提交的任务全部处理完再退出。
// 收尾处理使用后台 ctx（safeConsume 内部），不随原 ctx 取消而中断。
func (p *WorkerPool) drain(workerID int, shard <-chan Job) {
	for {
		select {
		case job := <-shard:
			p.consumeJob(job)
		default:
			return
		}
	}
}

// consumeJob 消费单个任务：handler 出错或 panic 时不提交 offset（消息将按
// at-least-once 语义重新投递），只有处理成功才调用提交回调。
func (p *WorkerPool) consumeJob(job Job) {
	if err := p.safeConsume(job.Msg); err != nil {
		appLogger.Errorf("consume failed, skip commit: %v", err)
		return
	}
	if job.Commit != nil {
		job.Commit()
	}
}

// safeConsume 包装 handler.Consume，把 panic 转成 error（带堆栈），保证 worker goroutine 不死亡。
func (p *WorkerPool) safeConsume(msg *Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v\n%s", r, debug.Stack())
		}
	}()
	return p.handler.Consume(context.Background(), msg)
}

// Submit 提交任务。按 Key 哈希路由到固定分片，同 key 必进同一分片（同一 worker 串行）。
// 当目标分片队列满时阻塞等待（形成天然背压，不会丢消息）。
func (p *WorkerPool) Submit(e *Event, commit CommitFunc) {
	_ = p.SubmitCtx(context.Background(), e, commit)
}

// SubmitCtx 提交任务并响应 ctx 取消。
// 分片队列满时阻塞等待（背压），ctx 取消或池已停止时返回错误，消息不会入队（未提交，可重投）。
func (p *WorkerPool) SubmitCtx(ctx context.Context, e *Event, commit CommitFunc) error {
	var key []byte
	if e != nil && e.Msg != nil {
		key = e.Msg.Key
	}
	select {
	case <-p.stopCh:
		return ErrPoolClosed
	default:
	}
	select {
	case p.shards[p.shardIndex(key)] <- Job{Msg: e, Commit: commit}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// StopAndWait 优雅关闭：停止接收新任务，等待已提交任务全部处理完毕。
func (p *WorkerPool) StopAndWait() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
	p.wg.Wait()
}

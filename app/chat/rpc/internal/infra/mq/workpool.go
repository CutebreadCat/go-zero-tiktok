package mq

import (
	"context"

	mqcontract "go_zero-tiktok/pkg/mq"
)

// CommitFunc 提交 Offset 的回调函数类型
// 由 Partition 层封装，WorkerPool 在消息处理成功后调用
type CommitFunc func()

// Job WorkerPool 的任务单元，包含业务消息和提交回调
type Job struct {
	Msg    *mqcontract.Event
	Commit CommitFunc
}

type WorkerPool struct {
	workers int
	queue   chan Job
	handler mqcontract.ConsumerHandler
}

func NewWorkerPool(workers int, queueSize int, h mqcontract.ConsumerHandler) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		queue:   make(chan Job, queueSize),
		handler: h,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		go func(workerID int) {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-p.queue:
					err := p.handler.Consume(ctx, job.Msg)
					if err != nil {
						continue
					}
					if job.Commit != nil {
						job.Commit()
					}
				}
			}
		}(i)
	}
}

// Submit 提交任务，包含业务消息和提交 Offset 的回调
func (p *WorkerPool) Submit(e *mqcontract.Event, commit CommitFunc) {
	p.queue <- Job{Msg: e, Commit: commit}
}

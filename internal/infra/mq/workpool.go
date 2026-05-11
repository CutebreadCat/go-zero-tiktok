package mq

import (
	"context"

	"go_zero-tiktok/internal/domain/mq"
)

type WorkerPool struct {
	workers int
	queue   chan *mq.Event
	handler mq.ConsumerHandler
}

func NewWorkerPool(workers int, queueSize int, h mq.ConsumerHandler) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		queue:   make(chan *mq.Event, queueSize),
		handler: h,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-p.queue:
					_ = p.handler.Consume(ctx, msg)
				}
			}
		}()
	}
}

func (p *WorkerPool) Submit(e *mq.Event) {
	p.queue <- e
}

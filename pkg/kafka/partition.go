package kafka

import (
	"context"
	"errors"
	"time"

	appLogger "go_zero-tiktok/pkg/logger"
)

// MessageReader 消息读取器接口，由 KafkaReader 实现
type MessageReader interface {
	Fetch(context.Context) (*Event, error)
	Commit(context.Context, *Message) error
}

type Partition struct {
	reader MessageReader
	pool   *WorkerPool
}

func NewPartition(reader MessageReader, pool *WorkerPool) *Partition {
	return &Partition{reader: reader, pool: pool}
}

func (p *Partition) Start(ctx context.Context) {
	appLogger.Info("partition fetcher starting")
	go func() {
		appLogger.Info("partition fetcher goroutine started")
		for {
			select {
			case <-ctx.Done():
				appLogger.Infof("partition fetcher stopped: %v", ctx.Err())
				return
			default:
			}
			event, err := p.reader.Fetch(ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				appLogger.Errorf("partition fetch failed: %v", err)
				time.Sleep(partitionFetchRetryWait)
				continue
			}
			msg := event.Msg
			p.pool.Submit(event, func() {
				if err := p.reader.Commit(ctx, msg); err != nil {
					appLogger.Errorf("commit message failed: %v", err)
				} else {
					appLogger.Infof("message committed: key=%s", string(msg.Key))
				}
			})
		}
	}()
}

package mykafka

import (
	"context"
	"errors"
	"time"

	appLogger "go_zero-tiktok/Prometheus/logger"
	"go_zero-tiktok/app/chat/rpc/internal/infra/mq"
	mqcontract "go_zero-tiktok/pkg/mq"
)

type Readder interface {
	Fetch(context.Context) (*mqcontract.Event, error)
	Commit(context.Context, *mqcontract.Message) error
}
type Partition struct {
	reader Readder
	pool   *mq.WorkerPool
}

func NewPartition(reader Readder, pool *mq.WorkerPool) *Partition {
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

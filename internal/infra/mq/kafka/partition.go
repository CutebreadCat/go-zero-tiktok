package mykafka

import (
	"context"
	"log"
	"time"

	"go_zero-tiktok/internal/infra/mq"

	mqcontract "go_zero-tiktok/internal/shared/mq"
)

// Readder 消息读取器接口，封装 Kafka Reader 的核心操作
type Readder interface {
	Fetch(ctx context.Context) (*mqcontract.Event, error)
	Commit(ctx context.Context, msg *mqcontract.Message) error
}

// Partition 单分区消费器，负责拉取消息并提交给 WorkerPool
type Partition struct {
	reader Readder
	pool   *mq.WorkerPool
}

func NewPartition(reader Readder, pool *mq.WorkerPool) *Partition {
	return &Partition{
		reader: reader,
		pool:   pool,
	}
}

func (p *Partition) Start(ctx context.Context) {
	log.Printf("🚀 Partition fetcher starting...")
	go func() {
		log.Printf("✅ Partition fetcher goroutine started")
		for {
			select {
			case <-ctx.Done():
				log.Printf("🛑 Partition fetcher stopped: %v", ctx.Err())
				return
			default:
				event, err := p.reader.Fetch(ctx)
				if err != nil {
					if err == context.DeadlineExceeded {
						continue
					}
					log.Printf("❌ Partition fetch error: %v", err)
					time.Sleep(2 * time.Second)
					continue
				}

				log.Printf("📨 Partition fetched message: key=%s, topic=%s", string(event.Msg.Key), event.Msg.Topic)

				msg := event.Msg
				commitFunc := func() {
					err := p.reader.Commit(ctx, msg)
					if err != nil {
						log.Printf("❌ Failed to commit message: %v", err)
					} else {
						log.Printf("✅ Message committed: key=%s", string(msg.Key))
					}
				}

				p.pool.Submit(event, commitFunc)
			}
		}
	}()
}

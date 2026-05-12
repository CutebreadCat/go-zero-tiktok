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
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				event, err := p.reader.Fetch(ctx)
				if err != nil {
					log.Printf("拉取消息失败: %v", err)
					time.Sleep(time.Second) // 避免疯狂重试
					continue
				}

				// 封装提交 Offset 的回调函数
				msg := event.Msg
				commitFunc := func() {
					if err := p.reader.Commit(ctx, msg); err != nil {
						log.Printf("提交 Offset 失败: %v", err)
					}
				}

				// 将业务消息和提交回调一起丢给 WorkerPool
				p.pool.Submit(event, commitFunc)
			}
		}
	}()
}

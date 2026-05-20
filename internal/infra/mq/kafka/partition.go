package mykafka

import (
	"context"
	"log"
	"time"

	"go_zero-tiktok/internal/infra/mq"

	mqcontract "go_zero-tiktok/internal/shared/mq"
)

type Readder interface {
	Fetch(ctx context.Context) (*mqcontract.Event, error)
	Commit(ctx context.Context, msg *mqcontract.Message) error
}

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
	log.Printf("Partition fetcher 启动...")
	go func() {
		log.Printf("Partition fetcher goroutine 已启动")
		for {
			select {
			case <-ctx.Done():
				log.Printf("Partition fetcher 停止: %v", ctx.Err())
				return
			default:
				event, err := p.reader.Fetch(ctx)
				if err != nil {
					if err == context.DeadlineExceeded {
						continue
					}
					log.Printf("Partition fetch 错误: %v", err)
					time.Sleep(2 * time.Second)
					continue
				}

				log.Printf("Partition 获取消息: key=%s, topic=%s", string(event.Msg.Key), event.Msg.Topic)

				msg := event.Msg
				commit_func := func() {
					err := p.reader.Commit(ctx, msg)
					if err != nil {
						log.Printf("提交消息失败: %v", err)
					} else {
						log.Printf("消息已提交: key=%s", string(msg.Key))
					}
				}

				p.pool.Submit(event, commit_func)
			}
		}
	}()
}

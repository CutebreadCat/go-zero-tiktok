package mykafka

import (
	"context"
	"fmt"
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
	fmt.Printf("Partition fetcher 启动...\n")
	go func() {
		fmt.Printf("Partition fetcher goroutine 已启动\n")
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Partition fetcher 停止: %v\n", ctx.Err())
				return
			default:
				event, err := p.reader.Fetch(ctx)
				if err != nil {
					if err == context.DeadlineExceeded {
						continue
					}
					fmt.Printf("Partition fetch 错误: %v\n", err)
					time.Sleep(partitionFetchRetryWait)
					continue
				}

				fmt.Printf("Partition 获取消息: key=%s, topic=%s\n", string(event.Msg.Key), event.Msg.Topic)

				msg := event.Msg
				commit_func := func() {
					err := p.reader.Commit(ctx, msg)
					if err != nil {
						fmt.Printf("提交消息失败: %v\n", err)
					} else {
						fmt.Printf("消息已提交: key=%s\n", string(msg.Key))
					}
				}

				p.pool.Submit(event, commit_func)
			}
		}
	}()
}

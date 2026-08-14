package kafka

import (
	"context"

	appLogger "go_zero-tiktok/pkg/logger"

	kafkaGo "github.com/segmentio/kafka-go"
)

type ConsumerTopicConfig struct{ Topic string }

type MultiTopicConsumerUnit struct {
	pool     *WorkerPool
	fetchers []*Partition
}

func NewMultiTopicConsumerUnit(pool *WorkerPool, fetchers []*Partition) *MultiTopicConsumerUnit {
	return &MultiTopicConsumerUnit{pool: pool, fetchers: fetchers}
}

// NewMultiTopicConsumerUnitFromConfigs 创建消费组模式的多 topic 消费单元。
// 每个 topic 一个 group reader，由 Kafka 消费组负责分区分配与负载均衡；
// 消息进入分片 WorkerPool，按 key 保序消费。
// retryCfg 可配置失败回调（OnFailure，如降级同步落库）、重试次数与 DLQ。
func NewMultiTopicConsumerUnitFromConfigs(configs []ConsumerTopicConfig, brokers []string, groupID string, handler ConsumerHandler, workerCount, queueSize int, retryCfg RetryConfig) *MultiTopicConsumerUnit {
	// 失败处理：有限次指数退避重试，超过上限先触发 OnFailure 降级回调（若配置），
	// 仍失败则写 DLQ（若配置）/跳过并提交 offset，
	// 防止单条 poison message 阻塞整个分片消费进度。
	pool := NewWorkerPool(workerCount, queueSize, NewRetryConsumer(handler, retryCfg))
	fetchers := make([]*Partition, 0, len(configs))
	for _, cfg := range configs {
		r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
			Brokers:        brokers,
			Topic:          cfg.Topic,
			GroupID:        groupID,
			MinBytes:       readerMinBytes,
			MaxBytes:       readerMaxBytes,
			MaxWait:        readerMaxWait,
			CommitInterval: 0, // 同步提交：offset 提交节奏完全由 commitCoordinator 控制
		})
		fetchers = append(fetchers, NewPartition(NewReader(r), pool))
	}
	return &MultiTopicConsumerUnit{pool: pool, fetchers: fetchers}
}

func (u *MultiTopicConsumerUnit) Start(ctx context.Context) {
	u.pool.Start(ctx)
	// 消费组模式：kafka-go 在首次 Fetch 时自动加入消费组并等待分区分配，
	// 无需猜测性的 stabilize sleep。
	for _, f := range u.fetchers {
		f.Start(ctx)
	}
}

// Stop 优雅关闭：先停止 fetcher（不再提交新任务），再等 pool 把已提交任务处理完，
// 最后关闭 reader（解除阻塞中的 Fetch 并归还消费组分区）。
func (u *MultiTopicConsumerUnit) Stop() {
	for _, f := range u.fetchers {
		f.Stop()
	}
	u.pool.StopAndWait()
	for _, f := range u.fetchers {
		if err := f.Close(); err != nil {
			appLogger.Errorf("close fetcher reader failed: %v", err)
		}
	}
}

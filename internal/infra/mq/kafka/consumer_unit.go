package mykafka

import (
	"context"

	"go_zero-tiktok/internal/infra/mq"
	mqcontract "go_zero-tiktok/internal/shared/mq"

	kafkaGo "github.com/segmentio/kafka-go"
)

// ConsumerTopicConfig 单个 topic 的消费配置
type ConsumerTopicConfig struct {
	Topic          string
	PartitionCount int
}

// MultiTopicConsumerUnit 核心编排器
// 将共享的 WorkerPool 和多个 PartitionFetcher 组合在一起，统一启动
type MultiTopicConsumerUnit struct {
	pool     *mq.WorkerPool
	fetchers []*Partition
}

func NewMultiTopicConsumerUnit(
	pool *mq.WorkerPool,
	fetchers []*Partition,
) *MultiTopicConsumerUnit {
	return &MultiTopicConsumerUnit{
		pool:     pool,
		fetchers: fetchers,
	}
}

// NewMultiTopicConsumerUnitFromConfigs 工厂方法：支持多个 topic 的消费配置
// 所有 topic 共享同一个 WorkerPool
func NewMultiTopicConsumerUnitFromConfigs(
	configs []ConsumerTopicConfig,
	groupID string,
	handler mqcontract.ConsumerHandler,
	workerCount int,
	queueSize int,
) *MultiTopicConsumerUnit {
	pool := mq.NewWorkerPool(workerCount, queueSize, handler)

	var fetchers []*Partition
	for _, cfg := range configs {
		for i := 0; i < cfg.PartitionCount; i++ {
			r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
				Brokers:   brokers,
				GroupID:   groupID,
				Topic:     cfg.Topic,
				Partition: i,
				MinBytes:  10e3, // 10KB
				MaxBytes:  10e6, // 10MB
			})
			reader := NewReader(r)
			fetchers = append(fetchers, NewPartition(reader, pool))
		}
	}

	return &MultiTopicConsumerUnit{
		pool:     pool,
		fetchers: fetchers,
	}
}

func (u *MultiTopicConsumerUnit) Start(ctx context.Context) {
	// 先启动 WorkerPool，准备好消费协程
	u.pool.Start(ctx)

	// 再启动所有 PartitionFetcher，开始拉取消息
	for _, f := range u.fetchers {
		f.Start(ctx)
	}
}

package mykafka

import (
	"context"
	"log"
	"time"

	"go_zero-tiktok/internal/infra/mq"
	mqcontract "go_zero-tiktok/internal/shared/mq"

	kafkaGo "github.com/segmentio/kafka-go"
)

type ConsumerTopicConfig struct {
	Topic string
}

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

func NewMultiTopicConsumerUnitFromConfigs(
	configs []ConsumerTopicConfig,
	brokers []string,
	groupID string,
	handler mqcontract.ConsumerHandler,
	workerCount int,
	queueSize int,
) *MultiTopicConsumerUnit {
	pool := mq.NewWorkerPool(workerCount, queueSize, handler)

	var fetchers []*Partition
	for _, cfg := range configs {
		r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
			Brokers:   brokers,
			Topic:     cfg.Topic,
			Partition: 0,
			MinBytes:  1,
			MaxBytes:  10e6,
			MaxWait:   500 * time.Millisecond,
		})
		log.Printf("Kafka Reader 配置（Partition 模式）：Brokers=%v, Topic=%s, Partition=0",
			brokers, cfg.Topic)
		reader := NewReader(r)
		fetchers = append(fetchers, NewPartition(reader, pool))
	}

	return &MultiTopicConsumerUnit{
		pool:     pool,
		fetchers: fetchers,
	}
}

func (u *MultiTopicConsumerUnit) Start(ctx context.Context) {
	log.Printf("MultiTopicConsumerUnit 启动，共 %d 个 fetcher", len(u.fetchers))

	// 先启动 WorkerPool，确保有 worker 可以处理消息
	u.pool.Start(ctx)
	log.Printf("WorkerPool 已启动")

	// 等待 Kafka 消费者组稳定（避免 rebalance 期间拉取失败）
	log.Printf("等待 Kafka 消费者组稳定...")
	time.Sleep(3 * time.Second)
	log.Printf("Kafka 消费者组已就绪")

	// 再启动所有 PartitionFetcher，开始拉取消息
	for i, f := range u.fetchers {
		log.Printf("启动 fetcher %d...", i)
		f.Start(ctx)
		log.Printf("Fetcher %d 已启动", i)

		// 每个 fetcher 之间间隔一下，避免同时连接 Kafka 造成压力
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("MultiTopicConsumerUnit 已启动")
}

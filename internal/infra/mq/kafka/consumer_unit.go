package mykafka

import (
	"context"
	"log"
	"time"

	"go_zero-tiktok/internal/infra/mq"
	mqcontract "go_zero-tiktok/internal/shared/mq"

	kafkaGo "github.com/segmentio/kafka-go"
)

// ConsumerTopicConfig 单个 topic 的消费配置
type ConsumerTopicConfig struct {
	Topic string
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
	brokers []string,
	groupID string,
	handler mqcontract.ConsumerHandler,
	workerCount int,
	queueSize int,
) *MultiTopicConsumerUnit {
	pool := mq.NewWorkerPool(workerCount, queueSize, handler)

	var fetchers []*Partition
	for _, cfg := range configs {
		// 不使用消费者组，直接指定 partition 0 来消费
		r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
			Brokers:   brokers,
			Topic:     cfg.Topic,
			Partition: 0,                      // 直接读取 partition 0
			MinBytes:  1,                      // 1 字节，立即返回
			MaxBytes:  10e6,                   // 10MB
			MaxWait:   500 * time.Millisecond, // 最大等待时间
		})
		log.Printf("📋 Kafka Reader 配置（Partition 模式）：Brokers=%v, Topic=%s, Partition=0",
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
	log.Printf("🚀 MultiTopicConsumerUnit starting with %d fetchers", len(u.fetchers))

	// 先启动 WorkerPool，准备好消费协程
	u.pool.Start(ctx)
	log.Printf("✅ WorkerPool started")

	// 关键修复：等待 Kafka 消费者组完成 Rebalance
	// 在 Docker 环境中，容器同时启动会导致 Kafka 协调器压力过大
	// 需要给 Kafka 足够的时间来完成消费者组的初始化和分区分配
	log.Printf("⏳ Waiting for Kafka consumer group to stabilize...")
	time.Sleep(3 * time.Second)
	log.Printf("✅ Kafka consumer group should be ready now")

	// 再启动所有 PartitionFetcher，开始拉取消息
	for i, f := range u.fetchers {
		log.Printf("🚀 Starting fetcher %d...", i)
		f.Start(ctx)
		log.Printf("✅ Fetcher %d started", i)

		// 每个 fetcher 之间间隔一下，避免同时连接 Kafka 造成压力
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("✅ MultiTopicConsumerUnit started")
}

package mykafka

import (
	"context"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
	appLogger "go_zero-tiktok/Prometheus/logger"
	"go_zero-tiktok/app/chat/rpc/internal/infra/mq"
	mqcontract "go_zero-tiktok/pkg/mq"
)

type ConsumerTopicConfig struct{ Topic string }
type MultiTopicConsumerUnit struct {
	pool     *mq.WorkerPool
	fetchers []*Partition
}

func NewMultiTopicConsumerUnit(pool *mq.WorkerPool, fetchers []*Partition) *MultiTopicConsumerUnit {
	return &MultiTopicConsumerUnit{pool: pool, fetchers: fetchers}
}

func NewMultiTopicConsumerUnitFromConfigs(configs []ConsumerTopicConfig, brokers []string, groupID string, handler mqcontract.ConsumerHandler, workerCount, queueSize int) *MultiTopicConsumerUnit {
	pool := mq.NewWorkerPool(workerCount, queueSize, handler)
	fetchers := make([]*Partition, 0, len(configs))
	for _, cfg := range configs {
		r := kafkaGo.NewReader(kafkaGo.ReaderConfig{Brokers: brokers, Topic: cfg.Topic, Partition: defaultPartition, MinBytes: readerMinBytes, MaxBytes: readerMaxBytes, MaxWait: readerMaxWait})
		appLogger.Infof("Kafka reader configured: brokers=%v topic=%s partition=0", brokers, cfg.Topic)
		fetchers = append(fetchers, NewPartition(NewReader(r), pool))
	}
	return &MultiTopicConsumerUnit{pool: pool, fetchers: fetchers}
}

func (u *MultiTopicConsumerUnit) Start(ctx context.Context) {
	appLogger.Infof("starting consumer unit with %d fetchers", len(u.fetchers))
	u.pool.Start(ctx)
	appLogger.Info("worker pool started")
	appLogger.Info("waiting for Kafka consumer group")
	time.Sleep(consumerGroupStabilizeWait)
	appLogger.Info("Kafka consumer group ready")
	for i, f := range u.fetchers {
		appLogger.Infof("starting fetcher %d", i)
		f.Start(ctx)
		time.Sleep(fetcherStartInterval)
	}
	appLogger.Info("consumer unit started")
}

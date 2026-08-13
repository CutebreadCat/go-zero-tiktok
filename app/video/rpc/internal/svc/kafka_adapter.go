package svc

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/domain/interaction"
	"go_zero-tiktok/pkg/kafka"
	appLogger "go_zero-tiktok/pkg/logger"
)

// kafkaLikeEventProducer Kafka 实现的 interaction.LikeEventProducer。
type kafkaLikeEventProducer struct {
	producer *kafka.Producer
	topic    string
}

// NewKafkaLikeEventProducer 创建 Kafka 点赞事件生产者。
func NewKafkaLikeEventProducer(brokers []string, topic string) *kafkaLikeEventProducer {
	if topic == "" {
		topic = interaction.DefaultLikeTopic
	}
	return &kafkaLikeEventProducer{
		producer: kafka.NewProducer(brokers, topic),
		topic:    topic,
	}
}

// WatchErrors 启动后台 goroutine 监听异步发送失败。
// 仅当 Producer 为 Async=true 模式时才有意义；失败仅做日志+告警，数据一致性由兜底对账保证。
func (p *kafkaLikeEventProducer) WatchErrors() {
	go func() {
		for err := range p.producer.Errors() {
			appLogger.Errorf("Kafka 异步点赞事件发送失败: %v", err)
			// TODO: 接入 Prometheus 指标 like_producer_error_total
		}
	}()
}

func (p *kafkaLikeEventProducer) Send(ctx context.Context, event *interaction.LikeEvent) error {
	if event == nil {
		return nil
	}
	return p.producer.SendMessage(ctx, event.ToKafkaEvent(p.topic))
}

func (p *kafkaLikeEventProducer) Close() error {
	return p.producer.Close()
}

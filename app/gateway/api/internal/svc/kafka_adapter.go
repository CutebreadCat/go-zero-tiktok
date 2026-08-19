package svc

import (
	"context"

	"go_zero-tiktok/pkg/event"
	"go_zero-tiktok/pkg/kafka"
	appLogger "go_zero-tiktok/pkg/logger"
)

// HotScoreRecalcProducer 热度分重算事件生产者接口。
// Gateway 通过该接口解耦与 video.rpc 的直接调用。
type HotScoreRecalcProducer interface {
	Send(ctx context.Context, videoID int64) error
	Close() error
}

// kafkaHotScoreRecalcProducer Kafka 实现的热度分重算事件生产者。
type kafkaHotScoreRecalcProducer struct {
	producer *kafka.Producer
	topic    string
}

// NewKafkaHotScoreRecalcProducer 创建 Kafka 热度分重算事件生产者。
func NewKafkaHotScoreRecalcProducer(brokers []string, topic string) *kafkaHotScoreRecalcProducer {
	if topic == "" {
		topic = event.DefaultHotScoreRecalcTopic
	}
	return &kafkaHotScoreRecalcProducer{
		producer: kafka.NewProducer(brokers, topic),
		topic:    topic,
	}
}

// WatchErrors 启动后台 goroutine 监听异步发送失败。
// 仅当 Producer 为 Async=true 模式时才有意义；失败仅做日志+告警。
func (p *kafkaHotScoreRecalcProducer) WatchErrors() {
	go func() {
		for err := range p.producer.Errors() {
			appLogger.Errorf("Kafka 热度分重算事件发送失败: %v", err)
			// TODO: 接入 Prometheus 指标 hot_score_producer_error_total
		}
	}()
}

func (p *kafkaHotScoreRecalcProducer) Send(ctx context.Context, videoID int64) error {
	if videoID == 0 {
		return nil
	}
	return p.producer.SendMessage(ctx, event.HotScoreRecalcEvent{VideoID: videoID}.ToKafkaEvent(p.topic))
}

func (p *kafkaHotScoreRecalcProducer) Close() error {
	return p.producer.Close()
}

package svc

import (
	"context"

	"go_zero-tiktok/pkg/event"
	"go_zero-tiktok/pkg/kafka"
	appLogger "go_zero-tiktok/pkg/logger"
)

// VideoVisitProducer 视频访问事件生产者接口。
// 用于将 Feed/搜索/作者列表等查询产生的访问流量异步化，避免在主链路写库。
type VideoVisitProducer interface {
	Send(ctx context.Context, videoID int64, delta int64) error
	Close() error
}

// kafkaVideoVisitProducer Kafka 实现的视频访问事件生产者。
type kafkaVideoVisitProducer struct {
	producer *kafka.Producer
	topic    string
}

// NewKafkaVideoVisitProducer 创建 Kafka 视频访问事件生产者。
func NewKafkaVideoVisitProducer(brokers []string, topic string) *kafkaVideoVisitProducer {
	if topic == "" {
		topic = event.DefaultVideoVisitTopic
	}
	return &kafkaVideoVisitProducer{
		producer: kafka.NewProducer(brokers, topic),
		topic:    topic,
	}
}

// WatchErrors 启动后台 goroutine 监听异步发送失败。
func (p *kafkaVideoVisitProducer) WatchErrors() {
	go func() {
		for err := range p.producer.Errors() {
			appLogger.Errorf("Kafka 视频访问事件发送失败: %v", err)
			// TODO: 接入 Prometheus 指标 video_visit_producer_error_total
		}
	}()
}

func (p *kafkaVideoVisitProducer) Send(ctx context.Context, videoID int64, delta int64) error {
	if videoID == 0 || delta == 0 {
		return nil
	}
	return p.producer.SendMessage(ctx, event.VideoVisitEvent{VideoID: videoID, Delta: delta}.ToKafkaEvent(p.topic))
}

func (p *kafkaVideoVisitProducer) Close() error {
	return p.producer.Close()
}

package svc

import (
	"context"
	"time"

	"go_zero-tiktok/app/gateway/api/internal/types"
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

// TrackingEventProducer 埋点事件生产者接口。
type TrackingEventProducer interface {
	SendBatch(ctx context.Context, userID int64, events []types.TrackingEventRequest) error
	Close() error
}

// kafkaTrackingEventProducer Kafka 实现的埋点事件生产者。
type kafkaTrackingEventProducer struct {
	producer *kafka.Producer
	topic    string
}

// NewKafkaTrackingEventProducer 创建 Kafka 埋点事件生产者。
func NewKafkaTrackingEventProducer(brokers []string, topic string) *kafkaTrackingEventProducer {
	if topic == "" {
		topic = event.DefaultTrackingTopic
	}
	return &kafkaTrackingEventProducer{
		producer: kafka.NewProducer(brokers, topic),
		topic:    topic,
	}
}

// WatchErrors 启动后台 goroutine 监听异步发送失败。
func (p *kafkaTrackingEventProducer) WatchErrors() {
	go func() {
		for err := range p.producer.Errors() {
			appLogger.Errorf("Kafka 埋点事件发送失败: %v", err)
		}
	}()
}

func (p *kafkaTrackingEventProducer) SendBatch(ctx context.Context, userID int64, events []types.TrackingEventRequest) error {
	if len(events) == 0 {
		return nil
	}
	for _, ev := range events {
		kafkaEvent := buildTrackingKafkaEvent(userID, ev, p.topic)
		if kafkaEvent == nil {
			continue
		}
		if err := p.producer.SendMessage(ctx, kafkaEvent); err != nil {
			return err
		}
	}
	return nil
}

func (p *kafkaTrackingEventProducer) Close() error {
	return p.producer.Close()
}

// buildTrackingKafkaEvent 把 Gateway 请求转换为 kafka.Event。
func buildTrackingKafkaEvent(userID int64, req types.TrackingEventRequest, topic string) *kafka.Event {
	base := event.TrackingEventBase{
		Timestamp: req.Timestamp,
		UserID:    userID,
		DeviceID:  "", // Gateway 暂不采集 device_id
		ClientIP:  "",
	}
	if base.Timestamp == 0 {
		base.Timestamp = time.Now().UnixMilli()
	}

	switch req.EventType {
	case event.ImpressionType:
		return event.ImpressionEvent{
			TrackingEventBase: base,
			VideoID:           req.VideoID,
			Scene:             req.Scene,
			RequestID:         req.RequestID,
			Position:          req.Position,
		}.ToKafkaEvent(topic)
	case event.PlayType:
		return event.PlayEvent{
			TrackingEventBase: base,
			VideoID:           req.VideoID,
			Scene:             req.Scene,
			RequestID:         req.RequestID,
			DurationMs:        req.DurationMs,
		}.ToKafkaEvent(topic)
	case event.ProgressType:
		return event.ProgressEvent{
			TrackingEventBase: base,
			VideoID:           req.VideoID,
			ProgressPct:       req.Progress,
			WatchMs:           req.WatchMs,
			DurationMs:        req.DurationMs,
		}.ToKafkaEvent(topic)
	case event.CompleteType:
		return event.CompleteEvent{
			TrackingEventBase: base,
			VideoID:           req.VideoID,
			WatchMs:           req.WatchMs,
			DurationMs:        req.DurationMs,
		}.ToKafkaEvent(topic)
	case event.CommentType:
		return event.CommentEvent{
			TrackingEventBase: base,
			VideoID:           req.VideoID,
			CommentID:         req.CommentID,
		}.ToKafkaEvent(topic)
	case event.FollowType:
		return event.FollowEvent{
			TrackingEventBase: base,
			TargetUserID:      req.FollowUserID,
			Action:            req.Action,
		}.ToKafkaEvent(topic)
	case event.ShareType:
		return event.ShareEvent{
			TrackingEventBase: base,
			VideoID:           req.VideoID,
			Channel:           req.Channel,
		}.ToKafkaEvent(topic)
	default:
		return nil
	}
}

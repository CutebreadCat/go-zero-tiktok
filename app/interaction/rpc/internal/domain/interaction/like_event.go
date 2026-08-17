package interaction

import (
	"context"
	"encoding/json"
	"fmt"

	"go_zero-tiktok/pkg/kafka"
)

const (
	// LikeEventType 点赞事件类型，注册到 kafka 事件工厂用于消费端反序列化。
	LikeEventType = "LikeEvent"
	// DefaultLikeTopic 默认点赞 topic。
	DefaultLikeTopic = "video-like-events"
)

// LikeAction 互动事件动作类型。
type LikeAction string

const (
	LikeActionLike           LikeAction = "like"
	LikeActionCancel         LikeAction = "cancel"
	LikeActionFavorite       LikeAction = "favorite"
	LikeActionCancelFavorite LikeAction = "cancel_favorite"
)

func init() {
	// 注册反序列化工厂，消费端拿到 Type=LikeEvent 时自动创建 *LikeEvent。
	kafka.RegisterEventFactory(LikeEventType, func() any { return &LikeEvent{} })
}

// LikeEvent 视频互动事件（点赞/取消点赞/收藏/取消收藏）。
// 事件是领域对象，定义在 domain 层，基础设施层通过 Kafka 传输。
type LikeEvent struct {
	UserID   int64      `json:"user_id"`
	VideoID  int64      `json:"video_id"`
	Action   LikeAction `json:"action"`
	ClientIP string     `json:"client_ip,omitempty"`
}

// ToKafkaEvent 将 LikeEvent 包装为 kafka.Event。
func (e *LikeEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.VideoID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: LikeEventType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

// LikeEventFromKafkaEvent 从 kafka.Event 解析出 *LikeEvent。
func LikeEventFromKafkaEvent(event *kafka.Event) (*LikeEvent, error) {
	if event == nil {
		return nil, fmt.Errorf("event is nil")
	}
	if event.Type != LikeEventType {
		return nil, fmt.Errorf("unexpected event type: %s", event.Type)
	}

	switch v := event.Data.(type) {
	case *LikeEvent:
		return v, nil
	case LikeEvent:
		return &v, nil
	case map[string]any:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal map data: %w", err)
		}
		var e LikeEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, fmt.Errorf("unmarshal LikeEvent: %w", err)
		}
		return &e, nil
	default:
		return nil, fmt.Errorf("LikeEvent data type invalid: %T", event.Data)
	}
}

// LikeEventProducer 点赞事件生产者接口，领域层只依赖接口，不依赖具体消息中间件。
type LikeEventProducer interface {
	Send(ctx context.Context, event *LikeEvent) error
	Close() error
}

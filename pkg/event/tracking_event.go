package event

import (
	"encoding/json"
	"fmt"

	"go_zero-tiktok/pkg/kafka"
)

const (
	// ImpressionType 曝光事件类型。
	ImpressionType = "VideoImpression"
	// PlayType 播放事件类型。
	PlayType = "VideoPlay"
	// ProgressType 播放进度事件类型。
	ProgressType = "VideoProgress"
	// CompleteType 播放完成事件类型。
	CompleteType = "VideoComplete"
	// CommentType 评论事件类型。
	CommentType = "VideoComment"
	// FollowType 关注事件类型。
	FollowType = "UserFollow"
	// ShareType 分享事件类型。
	ShareType = "VideoShare"

	// DefaultTrackingTopic 默认埋点事件 topic。
	// Phase 2 先统一到一个 topic，按 Type 字段区分事件类型，简化消费侧。
	DefaultTrackingTopic = "tracking-events"
)

func init() {
	kafka.RegisterEventFactory(ImpressionType, func() any { return &ImpressionEvent{} })
	kafka.RegisterEventFactory(PlayType, func() any { return &PlayEvent{} })
	kafka.RegisterEventFactory(ProgressType, func() any { return &ProgressEvent{} })
	kafka.RegisterEventFactory(CompleteType, func() any { return &CompleteEvent{} })
	kafka.RegisterEventFactory(CommentType, func() any { return &CommentEvent{} })
	kafka.RegisterEventFactory(FollowType, func() any { return &FollowEvent{} })
	kafka.RegisterEventFactory(ShareType, func() any { return &ShareEvent{} })
}

// TrackingEventBase 所有埋点事件的通用字段。
// VideoID 没有放在基础结构体里，因为 FollowEvent 不一定发生在视频场景下。
type TrackingEventBase struct {
	Timestamp int64  `json:"timestamp"`  // 事件产生时间，UnixMilli
	UserID    int64  `json:"user_id"`    // 0 表示未登录
	DeviceID  string `json:"device_id"`  // 设备标识，未登录用户用
	ClientIP  string `json:"client_ip"`
}

// ImpressionEvent 曝光事件：视频卡片出现在客户端可视区域时上报。
type ImpressionEvent struct {
	TrackingEventBase
	VideoID   int64  `json:"video_id"`
	Scene     string `json:"scene"`      // recommend / following / hot / timeline
	RequestID string `json:"request_id"` // 一次 Feed 请求标识
	Position  int32  `json:"position"`   // 在列表中的位置
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e ImpressionEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.UserID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: ImpressionType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

// PlayEvent 播放事件：用户点击播放时上报。
type PlayEvent struct {
	TrackingEventBase
	VideoID    int64  `json:"video_id"`
	Scene      string `json:"scene"`
	RequestID  string `json:"request_id"`
	DurationMs int64  `json:"duration_ms"` // 视频总时长
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e PlayEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.UserID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: PlayType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

// ProgressEvent 播放进度事件：播放过程中按进度点上报（如 25%/50%/75%）。
type ProgressEvent struct {
	TrackingEventBase
	VideoID     int64 `json:"video_id"`
	ProgressPct int32 `json:"progress_pct"` // 0~100
	WatchMs     int64 `json:"watch_ms"`
	DurationMs  int64 `json:"duration_ms"`
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e ProgressEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.UserID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: ProgressType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

// CompleteEvent 播放完成事件：视频自然播放完时上报。
type CompleteEvent struct {
	TrackingEventBase
	VideoID    int64 `json:"video_id"`
	WatchMs    int64 `json:"watch_ms"`
	DurationMs int64 `json:"duration_ms"`
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e CompleteEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.UserID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: CompleteType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

// CommentEvent 评论事件：用户发表评论时上报。
type CommentEvent struct {
	TrackingEventBase
	VideoID   int64 `json:"video_id"`
	CommentID int64 `json:"comment_id"`
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e CommentEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.UserID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: CommentType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

// FollowEvent 关注事件：用户关注/取消关注时上报。
type FollowEvent struct {
	TrackingEventBase
	TargetUserID int64  `json:"target_user_id"` // 被关注者
	Action       string `json:"action"`         // follow / unfollow
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e FollowEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.UserID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: FollowType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

// ShareEvent 分享事件：用户分享视频时上报。
type ShareEvent struct {
	TrackingEventBase
	VideoID int64  `json:"video_id"`
	Channel string `json:"channel"`
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e ShareEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.UserID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: ShareType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

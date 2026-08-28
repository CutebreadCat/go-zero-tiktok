package event

import (
	"encoding/json"
	"fmt"

	"go_zero-tiktok/pkg/kafka"
)

const (
	// HotScoreRecalcType 热度分重算事件类型。
	HotScoreRecalcType = "HotScoreRecalc"
	// DefaultHotScoreRecalcTopic 默认热度分重算 topic。
	DefaultHotScoreRecalcTopic = "video-hot-score-recalc"

	// VideoVisitType 视频访问事件类型。
	VideoVisitType = "VideoVisit"
	// DefaultVideoVisitTopic 默认视频访问 topic。
	DefaultVideoVisitTopic = "video-visit-events"
)

func init() {
	kafka.RegisterEventFactory(HotScoreRecalcType, func() any { return &HotScoreRecalcEvent{} })
	kafka.RegisterEventFactory(VideoVisitType, func() any { return &VideoVisitEvent{} })
}

// HotScoreRecalcEvent 触发对指定视频重新计算热度分。
// 由 Gateway 在点赞/收藏/评论等互动事件成功后发送，video.rpc 消费并执行重算。
type HotScoreRecalcEvent struct {
	VideoID int64 `json:"video_id"`
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e HotScoreRecalcEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.VideoID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: HotScoreRecalcType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

// VideoVisitEvent 视频被浏览事件，用于异步累加访问量并刷新热度分。
// 由 video.rpc 在 Feed/搜索/作者列表等查询路径发送，自身消费处理。
type VideoVisitEvent struct {
	VideoID int64 `json:"video_id"`
	Delta   int64 `json:"delta"`
}

// ToKafkaEvent 将事件序列化为 kafka.Event。
func (e VideoVisitEvent) ToKafkaEvent(topic string) *kafka.Event {
	key := []byte(fmt.Sprintf("%d", e.VideoID))
	payload, _ := json.Marshal(e)
	return &kafka.Event{
		Type: VideoVisitType,
		Msg: &kafka.Message{
			Topic: topic,
			Key:   key,
			Value: payload,
		},
		Data: e,
	}
}

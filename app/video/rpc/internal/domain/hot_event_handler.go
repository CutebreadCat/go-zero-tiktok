package domain

import (
	"context"
	"encoding/json"
	"fmt"

	"go_zero-tiktok/pkg/event"
	"go_zero-tiktok/pkg/kafka"
	appLogger "go_zero-tiktok/pkg/logger"
)

// HotScoreEventHandler 消费 Kafka 热度相关事件并驱动 video 领域服务处理。
// 支持事件：
//   - HotScoreRecalc：重新计算视频热度分（来自 Gateway 点赞/收藏/评论）。
//   - VideoVisit：累加访问量并刷新热度分（来自本服务查询路径）。
type HotScoreEventHandler struct {
	videoService *VideoService
}

// NewHotScoreEventHandler 创建热度事件处理器。
func NewHotScoreEventHandler(s *VideoService) *HotScoreEventHandler {
	return &HotScoreEventHandler{videoService: s}
}

func (h *HotScoreEventHandler) Consume(ctx context.Context, msg *kafka.Event) error {
	if msg == nil {
		return nil
	}

	switch msg.Type {
	case event.HotScoreRecalcType:
		e, err := hotScoreRecalcFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("HotScoreEventHandler parse HotScoreRecalc failed: %v", err)
			return nil
		}
		return h.videoService.RecalculateHotScore(ctx, e.VideoID)

	case event.VideoVisitType:
		e, err := videoVisitFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("HotScoreEventHandler parse VideoVisit failed: %v", err)
			return nil
		}
		return h.videoService.IncreaseVideoVisitCount(ctx, e.VideoID, e.Delta)

	default:
		appLogger.Warnf("HotScoreEventHandler unknown event type: %s", msg.Type)
		return nil
	}
}

// Compile-time check: HotScoreEventHandler 实现 kafka.ConsumerHandler。
var _ kafka.ConsumerHandler = (*HotScoreEventHandler)(nil)

func hotScoreRecalcFromKafkaEvent(ev *kafka.Event) (*event.HotScoreRecalcEvent, error) {
	if ev.Type != event.HotScoreRecalcType {
		return nil, fmt.Errorf("unexpected event type: %s", ev.Type)
	}
	switch v := ev.Data.(type) {
	case *event.HotScoreRecalcEvent:
		return v, nil
	case event.HotScoreRecalcEvent:
		return &v, nil
	case map[string]any:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal map data: %w", err)
		}
		var e event.HotScoreRecalcEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, fmt.Errorf("unmarshal HotScoreRecalcEvent: %w", err)
		}
		return &e, nil
	default:
		return nil, fmt.Errorf("HotScoreRecalcEvent data type invalid: %T", ev.Data)
	}
}

func videoVisitFromKafkaEvent(ev *kafka.Event) (*event.VideoVisitEvent, error) {
	if ev.Type != event.VideoVisitType {
		return nil, fmt.Errorf("unexpected event type: %s", ev.Type)
	}
	switch v := ev.Data.(type) {
	case *event.VideoVisitEvent:
		return v, nil
	case event.VideoVisitEvent:
		return &v, nil
	case map[string]any:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal map data: %w", err)
		}
		var e event.VideoVisitEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, fmt.Errorf("unmarshal VideoVisitEvent: %w", err)
		}
		return &e, nil
	default:
		return nil, fmt.Errorf("VideoVisitEvent data type invalid: %T", ev.Data)
	}
}

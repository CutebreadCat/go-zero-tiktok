package domain

import (
	"context"
	"encoding/json"
	"fmt"

	"go_zero-tiktok/pkg/event"
	"go_zero-tiktok/pkg/kafka"
	appLogger "go_zero-tiktok/pkg/logger"
)

// TrackingEventHandler 消费 Kafka 埋点事件并驱动视频领域服务处理。
// 支持事件：
//   - VideoImpression：记录用户曝光，写入 feed:seen 与 video_view_events。
//   - VideoPlay / VideoProgress / VideoComplete：记录播放行为，写入 video_view_events。
//   - VideoComment / UserFollow / VideoShare：预留，用于用户画像与热度重算。
type TrackingEventHandler struct {
	feedSeenRepo       IFeedSeenRepo
	videoViewEventRepo IVideoViewEventRepo
	videoService       *VideoService
}

// NewTrackingEventHandler 创建埋点事件处理器。
func NewTrackingEventHandler(
	feedSeenRepo IFeedSeenRepo,
	videoViewEventRepo IVideoViewEventRepo,
	videoService *VideoService,
) *TrackingEventHandler {
	return &TrackingEventHandler{
		feedSeenRepo:       feedSeenRepo,
		videoViewEventRepo: videoViewEventRepo,
		videoService:       videoService,
	}
}

func (h *TrackingEventHandler) Consume(ctx context.Context, msg *kafka.Event) error {
	if msg == nil {
		return nil
	}

	switch msg.Type {
	case event.ImpressionType:
		e, err := impressionFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("TrackingEventHandler parse Impression failed: %v", err)
			return nil
		}
		return h.handleImpression(ctx, e)

	case event.PlayType:
		e, err := playFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("TrackingEventHandler parse Play failed: %v", err)
			return nil
		}
		return h.handlePlay(ctx, e)

	case event.ProgressType:
		e, err := progressFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("TrackingEventHandler parse Progress failed: %v", err)
			return nil
		}
		return h.handleProgress(ctx, e)

	case event.CompleteType:
		e, err := completeFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("TrackingEventHandler parse Complete failed: %v", err)
			return nil
		}
		return h.handleComplete(ctx, e)

	case event.CommentType:
		e, err := commentFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("TrackingEventHandler parse Comment failed: %v", err)
			return nil
		}
		return h.handleComment(ctx, e)

	case event.FollowType:
		e, err := followFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("TrackingEventHandler parse Follow failed: %v", err)
			return nil
		}
		return h.handleFollow(ctx, e)

	case event.ShareType:
		e, err := shareFromKafkaEvent(msg)
		if err != nil {
			appLogger.Warnf("TrackingEventHandler parse Share failed: %v", err)
			return nil
		}
		return h.handleShare(ctx, e)

	default:
		appLogger.Warnf("TrackingEventHandler unknown event type: %s", msg.Type)
		return nil
	}
}

// Compile-time check: TrackingEventHandler 实现 kafka.ConsumerHandler。
var _ kafka.ConsumerHandler = (*TrackingEventHandler)(nil)

func (h *TrackingEventHandler) handleImpression(ctx context.Context, e *event.ImpressionEvent) error {
	if e.VideoID <= 0 {
		return nil
	}

	// 1. 写入 Redis 曝光记录（已登录用户）
	if h.feedSeenRepo != nil && e.UserID > 0 {
		if err := h.feedSeenRepo.MarkSeen(ctx, e.UserID, []int64{e.VideoID}); err != nil {
			appLogger.Errorf("mark seen failed for user %d video %d: %v", e.UserID, e.VideoID, err)
		}
	}

	// 2. 写入曝光明细（可选，用于后续分析）
	if h.videoViewEventRepo != nil {
		if err := h.videoViewEventRepo.CreateEvent(ctx, e.UserID, e.VideoID, e.Scene, e.RequestID, "exposed", 0, 0); err != nil {
			appLogger.Errorf("create impression event failed for user %d video %d: %v", e.UserID, e.VideoID, err)
		}
	}

	return nil
}

func (h *TrackingEventHandler) handlePlay(ctx context.Context, e *event.PlayEvent) error {
	if e.VideoID <= 0 || h.videoViewEventRepo == nil {
		return nil
	}
	return h.videoViewEventRepo.CreateEvent(ctx, e.UserID, e.VideoID, e.Scene, e.RequestID, "play", 0, 0)
}

func (h *TrackingEventHandler) handleProgress(ctx context.Context, e *event.ProgressEvent) error {
	if e.VideoID <= 0 || h.videoViewEventRepo == nil {
		return nil
	}
	return h.videoViewEventRepo.CreateEvent(ctx, e.UserID, e.VideoID, "", "", "progress", e.WatchMs, 0)
}

func (h *TrackingEventHandler) handleComplete(ctx context.Context, e *event.CompleteEvent) error {
	if e.VideoID <= 0 || h.videoViewEventRepo == nil {
		return nil
	}
	return h.videoViewEventRepo.CreateEvent(ctx, e.UserID, e.VideoID, "", "", "complete", e.WatchMs, 1)
}

func (h *TrackingEventHandler) handleComment(ctx context.Context, e *event.CommentEvent) error {
	if e.VideoID <= 0 {
		return nil
	}
	// Phase 2 预留：后续可更新用户画像、触发热度分重算。
	appLogger.Infof("tracking comment event: user=%d video=%d comment=%d", e.UserID, e.VideoID, e.CommentID)
	return nil
}

func (h *TrackingEventHandler) handleFollow(ctx context.Context, e *event.FollowEvent) error {
	// Phase 2 预留：后续可更新用户画像、社交关系。
	appLogger.Infof("tracking follow event: user=%d target=%d action=%s", e.UserID, e.TargetUserID, e.Action)
	return nil
}

func (h *TrackingEventHandler) handleShare(ctx context.Context, e *event.ShareEvent) error {
	if e.VideoID <= 0 {
		return nil
	}
	// Phase 2 预留：后续可更新用户画像。
	appLogger.Infof("tracking share event: user=%d video=%d channel=%s", e.UserID, e.VideoID, e.Channel)
	return nil
}

// IFeedSeenRepo 曝光记录仓库接口（避免 domain 层直接依赖 dal）。
type IFeedSeenRepo interface {
	MarkSeen(ctx context.Context, userID int64, videoIDs []int64) error
}

// IVideoViewEventRepo 视频浏览/播放事件仓库接口。
type IVideoViewEventRepo interface {
	CreateEvent(ctx context.Context, userID, videoID int64, scene, requestID, eventType string, watchMs int64, completed int8) error
}

func eventFromKafkaEvent[T any](ev *kafka.Event, eventType string, target T) (*T, error) {
	if ev.Type != eventType {
		return nil, fmt.Errorf("unexpected event type: %s", ev.Type)
	}
	switch v := ev.Data.(type) {
	case *T:
		return v, nil
	case T:
		return &v, nil
	case map[string]any:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal map data: %w", err)
		}
		var e T
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, fmt.Errorf("unmarshal %T: %w", target, err)
		}
		return &e, nil
	default:
		return nil, fmt.Errorf("event data type invalid: %T", ev.Data)
	}
}

func impressionFromKafkaEvent(ev *kafka.Event) (*event.ImpressionEvent, error) {
	return eventFromKafkaEvent(ev, event.ImpressionType, event.ImpressionEvent{})
}

func playFromKafkaEvent(ev *kafka.Event) (*event.PlayEvent, error) {
	return eventFromKafkaEvent(ev, event.PlayType, event.PlayEvent{})
}

func progressFromKafkaEvent(ev *kafka.Event) (*event.ProgressEvent, error) {
	return eventFromKafkaEvent(ev, event.ProgressType, event.ProgressEvent{})
}

func completeFromKafkaEvent(ev *kafka.Event) (*event.CompleteEvent, error) {
	return eventFromKafkaEvent(ev, event.CompleteType, event.CompleteEvent{})
}

func commentFromKafkaEvent(ev *kafka.Event) (*event.CommentEvent, error) {
	return eventFromKafkaEvent(ev, event.CommentType, event.CommentEvent{})
}

func followFromKafkaEvent(ev *kafka.Event) (*event.FollowEvent, error) {
	return eventFromKafkaEvent(ev, event.FollowType, event.FollowEvent{})
}

func shareFromKafkaEvent(ev *kafka.Event) (*event.ShareEvent, error) {
	return eventFromKafkaEvent(ev, event.ShareType, event.ShareEvent{})
}

package interaction

import (
	"context"

	"go_zero-tiktok/pkg/kafka"
	appLogger "go_zero-tiktok/pkg/logger"
)

// LikeDirtyMarker 标记视频互动状态需要被 flush 到 MySQL 的接口。
// Kafka 消费者不再直接写库，而是通过此接口把 video_id 加入脏集合，
// 由后台 LikeCountSyncer 统一批量同步到 MySQL，避免单条消息触发一次 DB 写。
type LikeDirtyMarker interface {
	MarkVideoLikeDirty(ctx context.Context, videoID int64) error
	MarkVideoFavoriteDirty(ctx context.Context, videoID int64) error
}

// LikeEventHandler 消费 Kafka 互动事件，仅把对应视频标记为脏，等待 syncer 批量落库。
// 注意：like 关系与计数已在请求路径写入 Redis；本 handler 不再直接操作 MySQL。
type LikeEventHandler struct {
	dirtyMarker LikeDirtyMarker
}

func NewLikeEventHandler(dirtyMarker LikeDirtyMarker) *LikeEventHandler {
	return &LikeEventHandler{dirtyMarker: dirtyMarker}
}

func (h *LikeEventHandler) Consume(ctx context.Context, event *kafka.Event) error {
	e, err := LikeEventFromKafkaEvent(event)
	if err != nil {
		appLogger.Warnf("LikeEventHandler parse event failed: %v", err)
		return nil
	}

	switch e.Action {
	case LikeActionLike, LikeActionCancel:
		return h.dirtyMarker.MarkVideoLikeDirty(ctx, e.VideoID)
	case LikeActionFavorite, LikeActionCancelFavorite:
		return h.dirtyMarker.MarkVideoFavoriteDirty(ctx, e.VideoID)
	default:
		appLogger.Warnf("LikeEventHandler unknown action: %s", e.Action)
		return nil
	}
}

// Compile-time check: LikeEventHandler 实现 kafka.ConsumerHandler。
var _ kafka.ConsumerHandler = (*LikeEventHandler)(nil)

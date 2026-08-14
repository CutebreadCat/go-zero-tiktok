package interaction

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/internal/cache"
	videodomain "go_zero-tiktok/app/interaction/rpc/internal/domain"
	appLogger "go_zero-tiktok/pkg/logger"
	"go_zero-tiktok/pkg/xerr"
)

// InteractionService 视频互动领域服务，负责点赞/收藏关系与计数。
// like_count 采用“先写 Redis + 发 Kafka 事件 + 定时批量 flush 到 MySQL”的最终一致性方案；
// favorite_count 仍保持同步，后续可同样异步化。
type InteractionService struct {
	interactionRepo videodomain.IVideoInteractionRepo
	popularRepo     videodomain.IPopularRepo
	likeCache       *cache.LikeCountCache
	eventProducer   LikeEventProducer
	asyncEnabled    bool
}

// NewInteractionService 创建互动领域服务。
// eventProducer 为 nil 时不发 Kafka，Redis 缓冲后由定时 syncer 刷回 MySQL（便于测试或 Kafka 未就绪）。
func NewInteractionService(
	interactionRepo videodomain.IVideoInteractionRepo,
	popularRepo videodomain.IPopularRepo,
	likeCache *cache.LikeCountCache,
	eventProducer LikeEventProducer,
) *InteractionService {
	asyncEnabled := eventProducer != nil
	if eventProducer == nil {
		eventProducer = noopLikeEventProducer{}
	}
	return &InteractionService{
		interactionRepo: interactionRepo,
		popularRepo:     popularRepo,
		likeCache:       likeCache,
		eventProducer:   eventProducer,
		asyncEnabled:    asyncEnabled,
	}
}

// LikeVideo 点赞视频：先写 Redis，再发 Kafka 事件。
func (s *InteractionService) LikeVideo(ctx context.Context, userID, videoID int64) error {
	if s.likeCache == nil {
		return s.syncLikeVideo(ctx, userID, videoID)
	}

	isNew, err := s.likeCache.LikeVideo(ctx, userID, videoID)
	if err != nil {
		return err
	}
	if !isNew {
		// 幂等：已经点过赞，直接返回业务重复错误。
		return xerr.NewInvalidParam("重复点赞")
	}

	if s.asyncEnabled {
		if err := s.publishLikeEvent(ctx, userID, videoID, LikeActionLike); err != nil {
			appLogger.Warnf("LikeVideo publish event failed: %v", err)
			// 事件发送失败不阻塞用户，丢给 syncer 兜底；如需强一致可改为 fallback。
		}
	}

	return nil
}

// CancelLikeVideo 取消点赞：先写 Redis，再发 Kafka 事件。
func (s *InteractionService) CancelLikeVideo(ctx context.Context, userID, videoID int64) error {
	if s.likeCache == nil {
		return s.syncCancelLikeVideo(ctx, userID, videoID)
	}

	cancelled, err := s.likeCache.CancelLikeVideo(ctx, userID, videoID)
	if err != nil {
		return err
	}
	if !cancelled {
		return xerr.NewInvalidParam("点赞关系不存在")
	}

	if s.asyncEnabled {
		if err := s.publishLikeEvent(ctx, userID, videoID, LikeActionCancel); err != nil {
			appLogger.Warnf("CancelLikeVideo publish event failed: %v", err)
		}
	}

	return nil
}

// syncLikeVideo Kafka/Redis 未就绪时的同步降级：直接写 MySQL。
func (s *InteractionService) syncLikeVideo(ctx context.Context, userID, videoID int64) error {
	if err := s.interactionRepo.LikeVideo(ctx, userID, videoID); err != nil {
		return err
	}
	return s.popularRepo.UpdateVideoLikeCount(ctx, videoID, 1)
}

// syncCancelLikeVideo Kafka/Redis 未就绪时的同步降级：直接写 MySQL。
func (s *InteractionService) syncCancelLikeVideo(ctx context.Context, userID, videoID int64) error {
	if err := s.interactionRepo.CancelLikeVideo(ctx, userID, videoID); err != nil {
		return err
	}
	return s.popularRepo.UpdateVideoLikeCount(ctx, videoID, -1)
}

// FavoriteVideo 收藏视频（计数仍走同步，后续可同样异步化）。
func (s *InteractionService) FavoriteVideo(ctx context.Context, userID, videoID int64) error {
	if err := s.interactionRepo.FavoriteVideo(ctx, userID, videoID); err != nil {
		return err
	}
	return s.popularRepo.UpdateVideoFavoriteCount(ctx, videoID, 1)
}

// CancelFavoriteVideo 取消收藏。
func (s *InteractionService) CancelFavoriteVideo(ctx context.Context, userID, videoID int64) error {
	if err := s.interactionRepo.CancelFavoriteVideo(ctx, userID, videoID); err != nil {
		return err
	}
	return s.popularRepo.UpdateVideoFavoriteCount(ctx, videoID, -1)
}

// GetLikedVideoIDs 获取用户点赞的视频 ID 列表（优先 Redis，未就绪则回源 MySQL）。
func (s *InteractionService) GetLikedVideoIDs(ctx context.Context, userID int64, pageNum, pageSize int32) ([]int64, int64, error) {
	if s.likeCache == nil {
		return s.interactionRepo.GetLikedVideoIDsByUserID(ctx, userID, pageNum, pageSize)
	}
	return s.likeCache.GetLikedVideoIDs(ctx, userID, pageNum, pageSize)
}

// GetFavoritedVideoIDs 获取用户收藏的视频 ID 列表。
func (s *InteractionService) GetFavoritedVideoIDs(ctx context.Context, userID int64, pageNum, pageSize int32) ([]int64, int64, error) {
	return s.interactionRepo.GetFavoritedVideoIDsByUserID(ctx, userID, pageNum, pageSize)
}

// GetLikeCounts 批量获取视频 like_count，优先读 Redis 缓存，未命中回源 MySQL。
func (s *InteractionService) GetLikeCounts(ctx context.Context, videoIDs []int64) (map[int64]int64, error) {
	if len(videoIDs) == 0 {
		return map[int64]int64{}, nil
	}

	if s.likeCache == nil {
		return s.popularRepo.GetLikeCounts(ctx, videoIDs)
	}

	cached, missed, err := s.likeCache.GetLikeCounts(ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	if len(missed) > 0 {
		stats, err := s.popularRepo.GetLikeCounts(ctx, missed)
		if err != nil {
			return nil, err
		}

		if err := s.likeCache.SetLikeCounts(ctx, stats); err != nil {
			appLogger.Warnf("GetLikeCounts write back to redis failed: %v", err)
		}

		for id, count := range stats {
			cached[id] = count
		}
	}

	return cached, nil
}

// GetFavoriteCounts 批量获取视频 favorite_count（当前直接读 MySQL，后续可同样加 Redis 缓存）。
func (s *InteractionService) GetFavoriteCounts(ctx context.Context, videoIDs []int64) (map[int64]int64, error) {
	if len(videoIDs) == 0 {
		return map[int64]int64{}, nil
	}
	return s.popularRepo.GetFavoriteCounts(ctx, videoIDs)
}

func (s *InteractionService) publishLikeEvent(ctx context.Context, userID, videoID int64, action LikeAction) error {
	return s.eventProducer.Send(ctx, &LikeEvent{
		UserID:  userID,
		VideoID: videoID,
		Action:  action,
	})
}

// noopLikeEventProducer 空实现，用于测试或 Kafka 未配置。
type noopLikeEventProducer struct{}

func (noopLikeEventProducer) Send(ctx context.Context, event *LikeEvent) error { return nil }
func (noopLikeEventProducer) Close() error                                     { return nil }

// Compile-time checks.
var _ LikeEventProducer = (noopLikeEventProducer{})

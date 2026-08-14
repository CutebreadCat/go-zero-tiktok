package worker

import (
	"context"
	"sync"
	"time"

	"go_zero-tiktok/app/interaction/rpc/internal/cache"
	videodomain "go_zero-tiktok/app/interaction/rpc/internal/domain"
	appLogger "go_zero-tiktok/pkg/logger"
)

const (
	defaultSyncInterval = 5 * time.Second
	defaultBatchSize    = 100
)

// LikeCountSyncer 后台定时器：将 Redis 中缓冲的点赞关系与计数批量 flush 到 MySQL。
// 核心策略：Redis 为事实来源，MySQL 为最终持久化；syncer 定期对齐两者。
type LikeCountSyncer struct {
	cache           *cache.LikeCountCache
	interactionRepo videodomain.IVideoInteractionRepo
	popularRepo     videodomain.IPopularRepo

	interval  time.Duration
	batchSize int

	stopOnce sync.Once
	stopCh   chan struct{}
}

// SyncerOption 配置函数。
type SyncerOption func(*LikeCountSyncer)

func WithSyncInterval(d time.Duration) SyncerOption {
	return func(s *LikeCountSyncer) { s.interval = d }
}

func WithBatchSize(n int) SyncerOption {
	return func(s *LikeCountSyncer) { s.batchSize = n }
}

// NewLikeCountSyncer 创建 syncer。
func NewLikeCountSyncer(
	likeCache *cache.LikeCountCache,
	interactionRepo videodomain.IVideoInteractionRepo,
	popularRepo videodomain.IPopularRepo,
	opts ...SyncerOption,
) *LikeCountSyncer {
	s := &LikeCountSyncer{
		cache:           likeCache,
		interactionRepo: interactionRepo,
		popularRepo:     popularRepo,
		interval:        defaultSyncInterval,
		batchSize:       defaultBatchSize,
		stopCh:          make(chan struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start 启动后台定时同步循环。
func (s *LikeCountSyncer) Start(ctx context.Context) {
	appLogger.Info("LikeCountSyncer starting")
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				appLogger.Infof("LikeCountSyncer stopped: %v", ctx.Err())
				return
			case <-s.stopCh:
				appLogger.Info("LikeCountSyncer stopped by Stop")
				return
			case <-ticker.C:
				if err := s.syncOnce(ctx); err != nil {
					appLogger.Errorf("LikeCountSyncer sync failed: %v", err)
				}
			}
		}
	}()
}

// Stop 优雅停止。
func (s *LikeCountSyncer) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}

// syncOnce 执行一次批量同步。
func (s *LikeCountSyncer) syncOnce(ctx context.Context) error {
	if s.cache == nil {
		return nil
	}

	videoIDs, err := s.cache.PopDirtyVideos(ctx, s.batchSize)
	if err != nil {
		return err
	}
	if len(videoIDs) == 0 {
		return nil
	}

	for _, videoID := range videoIDs {
		if err := s.syncVideo(ctx, videoID); err != nil {
			appLogger.Errorf("sync video %d failed: %v", videoID, err)
		}
	}

	return nil
}

// syncVideo 同步单个视频：关系 diff + 计数对齐。
func (s *LikeCountSyncer) syncVideo(ctx context.Context, videoID int64) error {
	// 1. 读取 Redis 当前关系。
	redisUserIDs, err := s.cache.GetVideoLikeUserIDs(ctx, videoID)
	if err != nil {
		return err
	}
	redisSet := toInt64Set(redisUserIDs)

	// 2. 读取 MySQL 当前关系。
	mysqlUserIDs, err := s.interactionRepo.GetLikeUserIDsByVideoID(ctx, videoID)
	if err != nil {
		return err
	}
	mysqlSet := toInt64Set(mysqlUserIDs)

	// 3. 计算差集。
	var toAdd, toRemove []int64
	for uid := range redisSet {
		if _, ok := mysqlSet[uid]; !ok {
			toAdd = append(toAdd, uid)
		}
	}
	for uid := range mysqlSet {
		if _, ok := redisSet[uid]; !ok {
			toRemove = append(toRemove, uid)
		}
	}

	// 4. 批量同步关系。
	if len(toAdd) > 0 {
		if err := s.interactionRepo.BatchAddLikeInteractions(ctx, videoID, toAdd); err != nil {
			appLogger.Errorf("batch add like interactions failed, video_id=%d: %v", videoID, err)
		}
	}
	if len(toRemove) > 0 {
		if err := s.interactionRepo.BatchRemoveLikeInteractions(ctx, videoID, toRemove); err != nil {
			appLogger.Errorf("batch remove like interactions failed, video_id=%d: %v", videoID, err)
		}
	}

	// 5. 对齐计数：以 Redis 用户集合大小为准。
	likeCount := int64(len(redisUserIDs))
	if err := s.popularRepo.SetLikeCount(ctx, videoID, likeCount); err != nil {
		// 如果视频 stat 记录不存在（如视频未发布），记录日志跳过。
		appLogger.Errorf("set like_count failed, video_id=%d count=%d: %v", videoID, likeCount, err)
	}

	// 6. 把 Redis count 基准同步为当前真实值（去除脏数据可能导致的累计偏差）。
	if err := s.cache.SetLikeCount(ctx, videoID, likeCount); err != nil {
		appLogger.Warnf("syncVideo set like_count cache failed, video_id=%d: %v", videoID, err)
	}

	return nil
}

func toInt64Set(ids []int64) map[int64]struct{} {
	set := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

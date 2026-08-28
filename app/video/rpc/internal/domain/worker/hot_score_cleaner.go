package worker

import (
	"context"
	"sync"
	"time"

	"go_zero-tiktok/app/video/rpc/internal/domain"
	appLogger "go_zero-tiktok/pkg/logger"
)

// HotScoreCleaner 定期清理过期热度分。
// 1. 删除超过滑动窗口未活跃的视频；
// 2. 裁剪热度榜，只保留 Top keepTopN。
type HotScoreCleaner struct {
	feedRepo  domain.IFeedRepo
	interval  time.Duration
	window    time.Duration
	keepTopN  int
	stopOnce  sync.Once
	stopCh    chan struct{}
}

// CleanerOption 配置函数。
type CleanerOption func(*HotScoreCleaner)

func WithCleanupInterval(d time.Duration) CleanerOption {
	return func(c *HotScoreCleaner) { c.interval = d }
}

func WithCleanupWindow(d time.Duration) CleanerOption {
	return func(c *HotScoreCleaner) { c.window = d }
}

func WithKeepTopN(n int) CleanerOption {
	return func(c *HotScoreCleaner) { c.keepTopN = n }
}

// NewHotScoreCleaner 创建清理器。
func NewHotScoreCleaner(feedRepo domain.IFeedRepo, opts ...CleanerOption) *HotScoreCleaner {
	c := &HotScoreCleaner{
		feedRepo: feedRepo,
		interval: 24 * time.Hour,
		window:   24 * time.Hour,
		keepTopN: 10000,
		stopCh:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Start 启动后台定时清理循环。
func (c *HotScoreCleaner) Start(ctx context.Context) {
	appLogger.Info("HotScoreCleaner starting")
	go func() {
		// 启动时立即执行一次清理
		if err := c.cleanupOnce(ctx); err != nil {
			appLogger.Errorf("HotScoreCleaner initial cleanup failed: %v", err)
		}

		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				appLogger.Infof("HotScoreCleaner stopped: %v", ctx.Err())
				return
			case <-c.stopCh:
				appLogger.Info("HotScoreCleaner stopped by Stop")
				return
			case <-ticker.C:
				if err := c.cleanupOnce(ctx); err != nil {
					appLogger.Errorf("HotScoreCleaner cleanup failed: %v", err)
				}
			}
		}
	}()
}

// Stop 优雅停止。
func (c *HotScoreCleaner) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

// cleanupOnce 执行一次清理：删除过期成员 + 裁剪规模。
func (c *HotScoreCleaner) cleanupOnce(ctx context.Context) error {
	if c.feedRepo == nil {
		return nil
	}

	cutoffMs := time.Now().Add(-c.window).UnixMilli()
	appLogger.Infof("HotScoreCleaner cleanup start, cutoff=%d, keepTopN=%d", cutoffMs, c.keepTopN)

	if err := c.feedRepo.CleanupExpiredHotVideos(ctx, cutoffMs, c.keepTopN); err != nil {
		return err
	}

	appLogger.Info("HotScoreCleaner cleanup done")
	return nil
}

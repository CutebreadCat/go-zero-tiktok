package worker

import (
	"context"
	"sync"
	"time"

	"go_zero-tiktok/app/video/rpc/internal/domain"
	appLogger "go_zero-tiktok/pkg/logger"
)

// HotScoreCleaner 定期裁剪热度榜规模，只保留 Top keepTopN。
// 榜单退出由排名决定；不再按"最后活跃时间"淘汰，避免低流量或停服超过窗口时长后整榜被清空。
type HotScoreCleaner struct {
	feedRepo domain.IFeedRepo
	interval time.Duration
	keepTopN int
	stopOnce sync.Once
	stopCh   chan struct{}
}

// CleanerOption 配置函数。
type CleanerOption func(*HotScoreCleaner)

func WithCleanupInterval(d time.Duration) CleanerOption {
	return func(c *HotScoreCleaner) { c.interval = d }
}

func WithKeepTopN(n int) CleanerOption {
	return func(c *HotScoreCleaner) { c.keepTopN = n }
}

// NewHotScoreCleaner 创建清理器。
func NewHotScoreCleaner(feedRepo domain.IFeedRepo, opts ...CleanerOption) *HotScoreCleaner {
	c := &HotScoreCleaner{
		feedRepo: feedRepo,
		interval: 24 * time.Hour,
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

// cleanupOnce 执行一次裁剪：只保留热度最高的 keepTopN 个成员。
func (c *HotScoreCleaner) cleanupOnce(ctx context.Context) error {
	if c.feedRepo == nil {
		return nil
	}

	appLogger.Infof("HotScoreCleaner cleanup start, keepTopN=%d", c.keepTopN)

	if err := c.feedRepo.TrimHotVideos(ctx, c.keepTopN); err != nil {
		return err
	}

	appLogger.Info("HotScoreCleaner cleanup done")
	return nil
}

package worker

import (
	"context"
	"sync"
	"time"

	"go_zero-tiktok/app/video/rpc/internal/domain"
	appLogger "go_zero-tiktok/pkg/logger"

	"github.com/zeromicro/go-zero/core/threading"
)

// HotScoreRefresher 定时重算榜单前 N 名的热度分。
//
// 热度分只在有事件时重算，写入后不会随时间自然衰减；本 worker 定期按当前时间
// 重新计算一次，让榜单持续逼近 HN 公式的真实值。只刷新已上榜成员，新成员仍由
// 实时重算链路写入。
type HotScoreRefresher struct {
	svc      *domain.VideoService
	feedRepo domain.IFeedRepo
	interval time.Duration
	topN     int
	stopOnce sync.Once
	stopCh   chan struct{}
}

// RefresherOption 配置函数。
type RefresherOption func(*HotScoreRefresher)

// WithRefreshInterval 设置刷新间隔。
func WithRefreshInterval(d time.Duration) RefresherOption {
	return func(r *HotScoreRefresher) { r.interval = d }
}

// WithRefreshTopN 设置每轮刷新的榜单条数。
func WithRefreshTopN(n int) RefresherOption {
	return func(r *HotScoreRefresher) { r.topN = n }
}

// NewHotScoreRefresher 创建热度分保鲜任务。
func NewHotScoreRefresher(svc *domain.VideoService, feedRepo domain.IFeedRepo, opts ...RefresherOption) *HotScoreRefresher {
	r := &HotScoreRefresher{
		svc:      svc,
		feedRepo: feedRepo,
		interval: 10 * time.Minute,
		topN:     10000,
		stopCh:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Start 启动后台刷新循环。
func (r *HotScoreRefresher) Start(ctx context.Context) {
	if r.svc == nil || r.interval <= 0 || r.topN <= 0 {
		return
	}

	// GoSafe 兜住 panic，避免定时任务把整个 RPC 进程带崩。
	threading.GoSafe(func() {
		appLogger.Infof("HotScoreRefresher starting: interval=%s topN=%d", r.interval, r.topN)

		// 启动后先等一个周期，避开服务启动与发布高峰。
		timer := time.NewTimer(r.interval)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				appLogger.Infof("HotScoreRefresher stopped: %v", ctx.Err())
				return
			case <-r.stopCh:
				appLogger.Info("HotScoreRefresher stopped by Stop")
				return
			case <-timer.C:
				r.refreshOnce(ctx)
				// 从本轮结束重新计时，任务超时也不会连续补跑。
				timer.Reset(r.interval)
			}
		}
	})
}

// Stop 优雅停止。
func (r *HotScoreRefresher) Stop() {
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
}

// refreshOnce 执行一轮分数刷新。
func (r *HotScoreRefresher) refreshOnce(ctx context.Context) {
	// 全量重建会用临时键 RENAME 覆盖正式榜，此时刷新属于白写，直接跳过。
	if r.feedRepo != nil {
		running, err := r.feedRepo.HotRebuildRunning(ctx)
		if err != nil {
			appLogger.Warnf("check hot rebuild lock failed: %v", err)
		} else if running {
			appLogger.Infof("skip hot ranking refresh: rebuild is running")
			return
		}
	}

	start := time.Now()
	refreshed, err := r.svc.RefreshHotRankingScores(ctx, r.topN)
	if err != nil {
		appLogger.Errorf("refresh hot ranking scores failed: %v", err)
		return
	}
	appLogger.Infof("HotScoreRefresher done: refreshed=%d cost=%s", refreshed, time.Since(start))
}

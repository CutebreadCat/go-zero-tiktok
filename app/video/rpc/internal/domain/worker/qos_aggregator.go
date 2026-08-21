package worker

import (
	"context"
	"strconv"
	"sync"
	"time"

	"go_zero-tiktok/app/video/rpc/internal/domain"
	"go_zero-tiktok/pkg/contract"
	appLogger "go_zero-tiktok/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// QoSAggregator 定期把 playback_qos_reports 原始上报聚合成 video_qos_stat 指标。
type QoSAggregator struct {
	qosRepo    domain.IPlaybackQoSRepo
	qosStatRepo domain.IVideoQoSRepo
	rdb        *redis.Redis
	interval   time.Duration
	batchSize  int32
	lastIDKey  string
	stopOnce   sync.Once
	stopCh     chan struct{}
}

// AggregatorOption 配置函数。
type AggregatorOption func(*QoSAggregator)

func WithAggregateInterval(d time.Duration) AggregatorOption {
	return func(a *QoSAggregator) { a.interval = d }
}

func WithAggregateBatchSize(n int32) AggregatorOption {
	return func(a *QoSAggregator) { a.batchSize = n }
}

func WithLastIDRedisKey(key string) AggregatorOption {
	return func(a *QoSAggregator) { a.lastIDKey = key }
}

// NewQoSAggregator 创建 QoS 聚合器。
func NewQoSAggregator(
	qosRepo domain.IPlaybackQoSRepo,
	qosStatRepo domain.IVideoQoSRepo,
	rdb *redis.Redis,
	opts ...AggregatorOption,
) *QoSAggregator {
	a := &QoSAggregator{
		qosRepo:     qosRepo,
		qosStatRepo: qosStatRepo,
		rdb:         rdb,
		interval:    5 * time.Minute,
		batchSize:   5000,
		lastIDKey:   "qos:aggregator:last_id",
		stopCh:      make(chan struct{}),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Start 启动后台定时聚合循环。
func (a *QoSAggregator) Start(ctx context.Context) {
	appLogger.Info("QoSAggregator starting")
	go func() {
		// 启动时立即执行一次聚合
		if err := a.aggregateOnce(ctx); err != nil {
			appLogger.Errorf("QoSAggregator initial aggregation failed: %v", err)
		}

		ticker := time.NewTicker(a.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				appLogger.Infof("QoSAggregator stopped: %v", ctx.Err())
				return
			case <-a.stopCh:
				appLogger.Info("QoSAggregator stopped by Stop")
				return
			case <-ticker.C:
				if err := a.aggregateOnce(ctx); err != nil {
					appLogger.Errorf("QoSAggregator aggregation failed: %v", err)
				}
			}
		}
	}()
}

// Stop 优雅停止。
func (a *QoSAggregator) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopCh)
	})
}

// aggregateOnce 执行一次增量聚合。
func (a *QoSAggregator) aggregateOnce(ctx context.Context) error {
	if a.qosRepo == nil || a.qosStatRepo == nil {
		return nil
	}

	lastID, err := a.loadLastID(ctx)
	if err != nil {
		return err
	}

	reports, err := a.qosRepo.GetReportsAfterID(ctx, lastID, a.batchSize)
	if err != nil {
		return err
	}
	if len(reports) == 0 {
		return nil
	}

	videoIDs := distinctVideoIDs(reports)
	allReports, err := a.qosRepo.GetReportsByVideoIDs(ctx, videoIDs)
	if err != nil {
		return err
	}

	metrics := aggregateMetrics(allReports)
	for vid, m := range metrics {
		if err := a.qosStatRepo.UpdateQoSAggregates(ctx, vid, m); err != nil {
			return err
		}
	}

	newLastID := maxReportID(reports)
	if err := a.saveLastID(ctx, newLastID); err != nil {
		logx.Errorf("QoSAggregator save last_id %d failed: %v", newLastID, err)
		return err
	}

	appLogger.Infof("QoSAggregator aggregated %d reports for %d videos, last_id=%d", len(reports), len(metrics), newLastID)
	return nil
}

func (a *QoSAggregator) loadLastID(ctx context.Context) (int64, error) {
	if a.rdb == nil {
		return 0, nil
	}
	val, err := a.rdb.GetCtx(ctx, a.lastIDKey)
	if err != nil {
		// Redis 未命中视为从 0 开始
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	id, _ := strconv.ParseInt(val, 10, 64)
	return id, nil
}

func (a *QoSAggregator) saveLastID(ctx context.Context, lastID int64) error {
	if a.rdb == nil {
		return nil
	}
	return a.rdb.SetexCtx(ctx, a.lastIDKey, strconv.FormatInt(lastID, 10), int(7*24*time.Hour.Seconds()))
}

func distinctVideoIDs(reports []*types.PlaybackQoSReport) []int64 {
	seen := make(map[int64]struct{}, len(reports))
	result := make([]int64, 0, len(reports))
	for _, r := range reports {
		if r == nil {
			continue
		}
		if _, ok := seen[r.VideoID]; ok {
			continue
		}
		seen[r.VideoID] = struct{}{}
		result = append(result, r.VideoID)
	}
	return result
}

func maxReportID(reports []*types.PlaybackQoSReport) int64 {
	var maxID int64
	for _, r := range reports {
		if r != nil && r.ID > maxID {
			maxID = r.ID
		}
	}
	return maxID
}

// aggregateMetrics 按 video_id 聚合 QoS 指标。
func aggregateMetrics(reports []*types.PlaybackQoSReport) map[int64]types.VideoQoSMetrics {
	groups := make(map[int64][]*types.PlaybackQoSReport)
	for _, r := range reports {
		if r == nil {
			continue
		}
		groups[r.VideoID] = append(groups[r.VideoID], r)
	}

	result := make(map[int64]types.VideoQoSMetrics, len(groups))
	for vid, list := range groups {
		result[vid] = calcVideoQoSMetrics(list)
	}
	return result
}

func calcVideoQoSMetrics(reports []*types.PlaybackQoSReport) types.VideoQoSMetrics {
	var (
		totalCount       int64
		completionSum    float64
		completionCount  int64
		stallReports     int64
		errorReports     int64
		bufferedMsSum    int64
		stallCountSum    int64
		bitrateSum       int64
		bitrateCount     int64
	)

	for _, r := range reports {
		if r == nil {
			continue
		}
		totalCount++

		if r.DurationMs > 0 && r.PlayedMs >= 0 {
			completionSum += float64(r.PlayedMs) / float64(r.DurationMs)
			completionCount++
		}

		if r.StallCount > 0 {
			stallReports++
		}
		stallCountSum += int64(r.StallCount)

		if r.ErrorCode != 0 {
			errorReports++
		}

		bufferedMsSum += r.BufferedMs

		if r.BitrateKbps > 0 {
			bitrateSum += int64(r.BitrateKbps)
			bitrateCount++
		}
	}

	if totalCount == 0 {
		return types.VideoQoSMetrics{}
	}

	m := types.VideoQoSMetrics{
		ReportCount:   totalCount,
		AvgBufferedMs: bufferedMsSum / totalCount,
		AvgStallCount: int32(stallCountSum / totalCount),
		StallRate:     int32(stallReports * 10000 / totalCount),
		ErrorRate:     int32(errorReports * 10000 / totalCount),
	}

	if completionCount > 0 {
		m.CompletionRate = int32(completionSum / float64(completionCount) * 10000)
	}
	if bitrateCount > 0 {
		m.AvgBitrateKbps = int32(bitrateSum / bitrateCount)
	}

	return m
}

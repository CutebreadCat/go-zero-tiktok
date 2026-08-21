package video_qos_stat

import (
	"context"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/xerr"

	"gorm.io/gorm"
)

// UpdateQoSAggregates 更新视频 QoS 聚合指标。
// 如果记录不存在则自动创建，保证发布后的视频能正常写入。
func UpdateQoSAggregates(ctx context.Context, db *gorm.DB, videoID int64, metrics types.VideoQoSMetrics) error {
	if videoID == 0 {
		return xerr.NewInvalidParam("视频 ID 不能为空")
	}

	record := &VideoQoSStat{
		VideoID:        videoID,
		CompletionRate: metrics.CompletionRate,
		StallRate:      metrics.StallRate,
		ErrorRate:      metrics.ErrorRate,
		AvgBitrateKbps: metrics.AvgBitrateKbps,
		AvgBufferedMs:  metrics.AvgBufferedMs,
		AvgStallCount:  metrics.AvgStallCount,
		ReportCount:    metrics.ReportCount,
	}

	if err := db.WithContext(ctx).
		Save(record).Error; err != nil {
		return xerr.Wrap(err, "update video qos aggregates failed")
	}
	return nil
}

// GetQoSMetricsByVideoIDs 批量查询视频 QoS 聚合指标。
func GetQoSMetricsByVideoIDs(ctx context.Context, db *gorm.DB, videoIDs []int64) (map[int64]types.VideoQoSMetrics, error) {
	if len(videoIDs) == 0 {
		return map[int64]types.VideoQoSMetrics{}, nil
	}

	var rows []VideoQoSStat
	if err := db.WithContext(ctx).Where("video_id IN ?", videoIDs).Find(&rows).Error; err != nil {
		return nil, xerr.Wrap(err, "get video qos metrics failed")
	}

	result := make(map[int64]types.VideoQoSMetrics, len(rows))
	for _, row := range rows {
		result[row.VideoID] = types.VideoQoSMetrics{
			CompletionRate: row.CompletionRate,
			StallRate:      row.StallRate,
			ErrorRate:      row.ErrorRate,
			AvgBitrateKbps: row.AvgBitrateKbps,
			AvgBufferedMs:  row.AvgBufferedMs,
			AvgStallCount:  row.AvgStallCount,
			ReportCount:    row.ReportCount,
		}
	}
	return result, nil
}

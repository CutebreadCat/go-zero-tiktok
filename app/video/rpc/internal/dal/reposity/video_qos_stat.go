package reposity

import (
	"context"

	qosstattable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_qos_stat"
	"go_zero-tiktok/pkg/contract"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

// VideoQoSRepo 视频播放质量聚合指标仓库。
type VideoQoSRepo struct {
	db *gorm.DB
}

// NewVideoQoSRepo 创建视频 QoS 聚合指标仓库。
func NewVideoQoSRepo(db *gorm.DB) *VideoQoSRepo {
	return &VideoQoSRepo{db: db}
}

// UpdateQoSAggregates 更新视频 QoS 聚合指标（不存在则创建）。
func (r *VideoQoSRepo) UpdateQoSAggregates(ctx context.Context, videoID int64, metrics types.VideoQoSMetrics) error {
	if err := qosstattable.UpdateQoSAggregates(ctx, r.db, videoID, metrics); err != nil {
		return pkgerrors.WithMessage(err, "VideoQoSRepo.UpdateQoSAggregates")
	}
	return nil
}

// GetQoSMetricsByVideoIDs 批量查询视频 QoS 聚合指标。
func (r *VideoQoSRepo) GetQoSMetricsByVideoIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoQoSMetrics, error) {
	metrics, err := qosstattable.GetQoSMetricsByVideoIDs(ctx, r.db, videoIDs)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "VideoQoSRepo.GetQoSMetricsByVideoIDs")
	}
	return metrics, nil
}

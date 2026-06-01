package video_popular

import (
	"context"

	"go_zero-tiktok/internal/shared/xerr"

	"gorm.io/gorm"
)

func IncreaseVideoVisitCount(ctx context.Context, db *gorm.DB, videoID string, delta int64) error {
	if delta <= 0 {
		delta = 1
	}

	result := db.WithContext(ctx).
		Model(&VideoPopular{}).
		Where("video_id = ?", videoID).
		Update("visit_count", gorm.Expr("visit_count + ?", delta))
	if result.Error != nil {
		return xerr.Wrap(result.Error, "increase video visit count failed")
	}

	return nil
}

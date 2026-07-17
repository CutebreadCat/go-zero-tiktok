package video_popular

import (
	"context"
	"fmt"

	"go_zero-tiktok/app/video/rpc/internal/dal/query"
	"go_zero-tiktok/pkg/xerr"

	"gorm.io/gorm"
)

func CreatePopularVideo(ctx context.Context, db *gorm.DB, videoID string) error {
	record := &VideoPopular{
		VideoID:      videoID,
		VisitCount:   0,
		LikeCount:    0,
		CommentCount: 0,
	}

	if err := db.WithContext(ctx).Create(record).Error; err != nil {
		return xerr.Wrap(err, "create popular video failed")
	}

	return nil
}

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

func UpdateVideoLikeCount(ctx context.Context, db *gorm.DB, videoID string, delta int64) error {
	result := db.WithContext(ctx).
		Model(&VideoPopular{}).
		Where("video_id = ?", videoID).
		Update("like_count", gorm.Expr("CASE WHEN like_count + ? < 0 THEN 0 ELSE like_count + ? END", delta, delta))
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update video like count failed")
	}

	if result.RowsAffected == 0 {
		return xerr.Wrap(fmt.Errorf("video %s not found", videoID), "update video like count failed")
	}

	return nil
}

func GetPopularVideoIDsByVisitCount(ctx context.Context, db *gorm.DB, pageNum, pageSize int32) ([]VideoPopular, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&VideoPopular{})

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get popular video count failed")
	}

	var rows []VideoPopular
	if err := dbQuery.Order("visit_count DESC").Scopes(query.Paginate(int(pageNum), int(pageSize))).
		Find(&rows).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get popular videos failed")
	}

	return rows, total, nil
}

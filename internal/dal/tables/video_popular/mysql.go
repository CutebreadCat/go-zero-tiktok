package video_popular

import (
	"context"
	"errors"
	"fmt"

	"go_zero-tiktok/internal/shared/xerr"

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
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&VideoPopular{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get popular video count failed")
	}

	var rows []VideoPopular
	offset := (pageNum - 1) * pageSize
	if err := query.Order("visit_count DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&rows).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get popular videos failed")
	}

	return rows, total, nil
}

func IncrVideoVisitCountInDB(ctx context.Context, db *gorm.DB, videoID string) error {
	result := db.WithContext(ctx).
		Model(&VideoPopular{}).
		Where("video_id = ?", videoID).
		Update("visit_count", gorm.Expr("visit_count + 1"))
	if result.Error != nil {
		return xerr.Wrap(result.Error, "incr video visit count failed")
	}

	return nil
}

func GetVideoPopularByVideoID(ctx context.Context, db *gorm.DB, videoID string) (*VideoPopular, error) {
	var videoPopular VideoPopular
	if err := db.WithContext(ctx).Where("video_id = ?", videoID).First(&videoPopular).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, "get video popular failed")
	}
	return &videoPopular, nil
}

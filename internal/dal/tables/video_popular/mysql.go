package video_popular

import (
	"context"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func CreatePopularVideo(ctx context.Context, db *gorm.DB, videoID string) error {
	logger := logx.WithContext(ctx)

	record := &types.VideoPopular{
		VideoID:      videoID,
		VisitCount:   0,
		LikeCount:    0,
		CommentCount: 0,
	}

	if err := db.WithContext(ctx).Create(record).Error; err != nil {
		logger.Errorf("create popular video failed: %v", err)
		return xerr.New(1002, "创建热门视频记录失败")
	}

	return nil
}

func IncreaseVideoVisitCount(ctx context.Context, db *gorm.DB, videoID string, delta int64) error {
	logger := logx.WithContext(ctx)

	if delta <= 0 {
		delta = 1
	}

	result := db.WithContext(ctx).
		Model(&types.VideoPopular{}).
		Where("video_id = ?", videoID).
		Update("visit_count", gorm.Expr("visit_count + ?", delta))
	if result.Error != nil {
		logger.Errorf("increase video visit count failed: %v", result.Error)
		return xerr.New(1002, "数据库当中增加视频访问次数失败")
	}

	if result.RowsAffected == 0 {
		logger.Errorf("video %s not found in DB, count lost.", videoID)
	}

	return nil
}

func UpdateVideoLikeCount(ctx context.Context, db *gorm.DB, videoID string, delta int64) error {
	logger := logx.WithContext(ctx)

	result := db.WithContext(ctx).
		Model(&types.VideoPopular{}).
		Where("video_id = ?", videoID).
		Update("like_count", gorm.Expr("CASE WHEN like_count + ? < 0 THEN 0 ELSE like_count + ? END", delta, delta))
	if result.Error != nil {
		logger.Errorf("update video like count failed: %v", result.Error)
		return xerr.New(1002, "更新视频点赞数失败")
	}

	if result.RowsAffected == 0 {
		logger.Errorf("update video like count failed: %v", gorm.ErrRecordNotFound)
		return xerr.New(1002, "更新视频点赞数失败")
	}

	return nil
}

func GetPopularVideoIDsByVisitCount(ctx context.Context, db *gorm.DB, pageNum, pageSize int32) ([]types.VideoPopular, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.VideoPopular{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get popular video count failed: %v", err)
		return nil, 0, xerr.New(1002, "获取热门视频总数失败")
	}

	var rows []types.VideoPopular
	offset := (pageNum - 1) * pageSize
	if err := query.Order("visit_count DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&rows).Error; err != nil {
		logger.Errorf("get popular videos failed: %v", err)
		return nil, 0, xerr.New(1002, "获取热门视频失败")
	}

	return rows, total, nil
}

func IncrVideoVisitCountInDB(ctx context.Context, db *gorm.DB, videoID string) error {
	logger := logx.WithContext(ctx)

	result := db.WithContext(ctx).
		Model(&types.VideoPopular{}).
		Where("video_id = ?", videoID).
		Update("visit_count", gorm.Expr("visit_count + 1"))
	if result.Error != nil {
		logger.Errorf("incr video visit count failed: %v", result.Error)
		return xerr.New(1002, "数据库当中增加视频访问次数失败")
	}

	if result.RowsAffected == 0 {
		logger.Errorf("video %s not found in DB, count lost.", videoID)
	}

	return nil
}

func GetVideoPopularByVideoID(ctx context.Context, db *gorm.DB, videoID string) (*types.VideoPopular, error) {
	logger := logx.WithContext(ctx)
	var videoPopular types.VideoPopular
	if err := db.WithContext(ctx).Where("video_id = ?", videoID).First(&videoPopular).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Errorf("video %s not found in DB", videoID)
			return nil, nil
		}
		logger.Errorf("get video popular failed: %v", err)
		return nil, xerr.New(1002, "获取视频热门信息失败")
	}
	return &videoPopular, nil
}

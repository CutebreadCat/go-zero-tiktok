package video_liker

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func LikeVideo(ctx context.Context, db *gorm.DB, userID, videoID string) error {
	logger := logx.WithContext(ctx)

	if userID == "" || videoID == "" {
		err := errors.New("userID or videoID is empty")
		logger.Errorf("like video failed: %v", err)
		return xerr.New(400, "用户ID或视频ID为空")
	}

	like := &types.VideoLiker{
		UserID:  userID,
		VideoID: videoID,
	}

	var existed types.VideoLiker
	err := db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).First(&existed).Error
	if err == nil {
		return xerr.New(400, "重复点赞")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("check like relation failed: %v", err)
		return xerr.New(1002, "查询点赞关系失败")
	}

	if err := db.WithContext(ctx).Create(like).Error; err != nil {
		logger.Errorf("like video failed: %v", err)
		return xerr.New(1002, "点赞视频失败")
	}

	return nil
}

func CancelLikeVideo(ctx context.Context, db *gorm.DB, userID, videoID string) error {
	logger := logx.WithContext(ctx)

	result := db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&types.VideoLiker{})
	if result.Error != nil {
		logger.Errorf("cancel like video failed: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("cancel like video failed: %v", err)
		return xerr.New(400, "点赞关系不存在")
	}

	return nil
}

func GetLikedVideoIDsByUserID(ctx context.Context, db *gorm.DB, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.VideoLiker{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get liked video ids count failed: %v", err)
		return nil, 0, xerr.New(1002, "获取点赞视频ID总数失败")
	}

	var likerRows []types.VideoLiker
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&likerRows).Error; err != nil {
		logger.Errorf("get liked video ids failed: %v", err)
		return nil, 0, xerr.New(1002, "获取点赞视频ID失败")
	}

	videoIDs := make([]string, 0, len(likerRows))
	for _, row := range likerRows {
		videoIDs = append(videoIDs, row.VideoID)
	}

	return videoIDs, total, nil
}

func GetAllLikedVideoIDsByUserID(ctx context.Context, db *gorm.DB, userID string) ([]string, error) {
	logger := logx.WithContext(ctx)

	var likerRows []types.VideoLiker
	if err := db.WithContext(ctx).Model(&types.VideoLiker{}).Where("user_id = ?", userID).Find(&likerRows).Error; err != nil {
		logger.Errorf("get all liked video ids failed: %v", err)
		return nil, xerr.New(1002, "获取全部点赞视频ID失败")
	}

	videoIDs := make([]string, 0, len(likerRows))
	for _, row := range likerRows {
		videoIDs = append(videoIDs, row.VideoID)
	}

	return videoIDs, nil
}

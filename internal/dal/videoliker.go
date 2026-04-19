package dal

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func LikeVideo(ctx context.Context, userID, videoID string) error {
	logger := logx.WithContext(ctx)

	if userID == "" || videoID == "" {
		err := errors.New("userID or videoID is empty")
		logger.Errorf("like video failed: %v", err)
		return err
	}

	like := &types.VideoLiker{
		UserID:  userID,
		VideoID: videoID,
	}

	if err := Db.WithContext(ctx).Create(like).Error; err != nil {
		logger.Errorf("like video failed: %v", err)
		return err
	}

	return nil
}

func CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	logger := logx.WithContext(ctx)

	result := Db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&types.VideoLiker{})
	if result.Error != nil {
		logger.Errorf("cancel like video failed: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("cancel like video failed: %v", err)
		return err
	}

	return nil
}

func GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := Db.WithContext(ctx).Model(&types.VideoLiker{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get liked video ids count failed: %v", err)
		return nil, 0, err
	}

	var likerRows []types.VideoLiker
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&likerRows).Error; err != nil {
		logger.Errorf("get liked video ids failed: %v", err)
		return nil, 0, err
	}

	videoIDs := make([]string, 0, len(likerRows))
	for _, row := range likerRows {
		videoIDs = append(videoIDs, row.VideoID)
	}

	return videoIDs, total, nil
}

package dal

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func CreateUserFollow(ctx context.Context, followerID, userID string) error {
	logger := logx.WithContext(ctx)

	if followerID == "" || userID == "" {
		err := errors.New("followerID or userID is empty")
		logger.Errorf("create user follow failed: %v", err)
		return err
	}

	relation := &types.UserFollow{
		FollowerID: followerID,
		UserID:     userID,
		Status:     0,
	}

	if err := Db.WithContext(ctx).Create(relation).Error; err != nil {
		logger.Errorf("create user follow failed: %v", err)
		return err
	}

	return nil
}

func UpdateUserFollowStatus(ctx context.Context, followerID, userID string, status int32) error {
	logger := logx.WithContext(ctx)

	result := Db.WithContext(ctx).
		Model(&types.UserFollow{}).
		Where("follower_id = ? AND user_id = ?", followerID, userID).
		Update("status", status)
	if result.Error != nil {
		logger.Errorf("update user follow status failed: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("update user follow status failed: %v", err)
		return err
	}

	return nil
}

func GetFollowingByFollowerID(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := Db.WithContext(ctx).Model(&types.UserFollow{}).Where("follower_id = ?", followerID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("count following list failed: %v", err)
		return nil, 0, err
	}

	var relations []types.UserFollow
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&relations).Error; err != nil {
		logger.Errorf("get following list failed: %v", err)
		return nil, 0, err
	}

	return relations, total, nil
}

func GetFansByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := Db.WithContext(ctx).Model(&types.UserFollow{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("count fans list failed: %v", err)
		return nil, 0, err
	}

	var relations []types.UserFollow
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&relations).Error; err != nil {
		logger.Errorf("get fans list failed: %v", err)
		return nil, 0, err
	}

	return relations, total, nil
}

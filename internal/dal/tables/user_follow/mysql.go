package user_follow

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func CreateUserFollow(ctx context.Context, db *gorm.DB, followerID, userID string) error {
	logger := logx.WithContext(ctx)

	if followerID == "" || userID == "" {
		err := errors.New("followerID or userID is empty")
		logger.Errorf("create user follow failed: %v", err)
		return xerr.New(400, "关注者ID或用户ID为空")
	}

	relation := &types.UserFollow{
		FollowerID: followerID,
		UserID:     userID,
		Status:     0,
	}

	if err := db.WithContext(ctx).Create(relation).Error; err != nil {
		logger.Errorf("create user follow failed: %v", err)
		return xerr.New(1002, "创建用户关注关系失败")
	}

	return nil
}

func UpdateUserFollowStatus(ctx context.Context, db *gorm.DB, followerID, userID string, status int32) error {
	logger := logx.WithContext(ctx)

	result := db.WithContext(ctx).
		Model(&types.UserFollow{}).
		Where("follower_id = ? AND user_id = ?", followerID, userID).
		Update("status", status)
	if result.Error != nil {
		logger.Errorf("update user follow status failed: %v", result.Error)
		return xerr.New(1002, "更新用户关注状态失败")
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("update user follow status failed: %v", err)
		return xerr.New(400, "用户关注关系不存在")
	}

	return nil
}

func GetFollowingISSubriber(ctx context.Context, db *gorm.DB, followerID, userID string) (bool, error) {
	logger := logx.WithContext(ctx)

	var relation types.UserFollow
	err := db.WithContext(ctx).Where("follower_id = ? AND user_id = ?", followerID, userID).First(&relation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		logger.Errorf("get following status failed: %v", err)
		return false, xerr.New(1002, "获取关注状态失败")
	}
	return true, nil
}

func GetFollowingByFollowerID(ctx context.Context, db *gorm.DB, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.UserFollow{}).Where("follower_id = ?", followerID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("count following list failed: %v", err)
		return nil, 0, xerr.New(1002, "统计关注列表失败")
	}

	var relations []types.UserFollow
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&relations).Error; err != nil {
		logger.Errorf("get following list failed: %v", err)
		return nil, 0, xerr.New(1002, "获取关注列表失败")
	}

	return relations, total, nil
}

func GetFansByUserID(ctx context.Context, db *gorm.DB, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.UserFollow{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("count fans list failed: %v", err)
		return nil, 0, xerr.New(1002, "统计粉丝列表失败")
	}

	var relations []types.UserFollow
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&relations).Error; err != nil {
		logger.Errorf("get fans list failed: %v", err)
		return nil, 0, xerr.New(1002, "获取粉丝列表失败")
	}

	return relations, total, nil
}

func GetFriendByUserID(ctx context.Context, db *gorm.DB, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.UserFollow{}).Where("user_id = ? AND status = ?", userID, 1)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("count friends list failed: %v", err)
		return nil, 0, xerr.New(1002, "统计好友列表失败")
	}

	var relations []types.UserFollow
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&relations).Error; err != nil {
		logger.Errorf("get friends list failed: %v", err)
		return nil, 0, xerr.New(1002, "获取好友列表失败")
	}

	return relations, total, nil
}

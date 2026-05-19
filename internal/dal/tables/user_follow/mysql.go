package user_follow

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/svc/xerr"

	"gorm.io/gorm"
)

func FollowUser(ctx context.Context, db *gorm.DB, followerID, userID string) error {
	if followerID == "" || userID == "" {
		return xerr.NewInvalidParam("关注者ID或用户ID为空")
	}
	if followerID == userID {
		return xerr.NewInvalidParam("不能关注自己")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var relation UserFollow
		err := tx.Where("follower_id = ? AND user_id = ?", followerID, userID).First(&relation).Error
		if err == nil {
			return xerr.NewInvalidParam("重复关注")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return xerr.Wrap(err, "query follow relation failed")
		}

		mutual := false
		err = tx.Where("follower_id = ? AND user_id = ?", userID, followerID).First(&relation).Error
		if err == nil {
			mutual = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return xerr.Wrap(err, "query reverse follow relation failed")
		}

		status := int32(0)
		if mutual {
			status = 1
		}

		if err := tx.Create(&UserFollow{FollowerID: followerID, UserID: userID, Status: status}).Error; err != nil {
			return xerr.Wrap(err, "create follow relation failed")
		}

		if mutual {
			if err := tx.Model(&UserFollow{}).
				Where("follower_id = ? AND user_id = ?", userID, followerID).
				Update("status", 1).Error; err != nil {
				return xerr.Wrap(err, "update reverse follow status failed")
			}
		}

		return nil
	})
}

func UnfollowUser(ctx context.Context, db *gorm.DB, followerID, userID string) error {
	if followerID == "" || userID == "" {
		return xerr.NewInvalidParam("关注者ID或用户ID为空")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("follower_id = ? AND user_id = ?", followerID, userID).Delete(&UserFollow{})
		if result.Error != nil {
			return xerr.Wrap(result.Error, "delete follow relation failed")
		}
		if result.RowsAffected == 0 {
			return xerr.NewInvalidParam("未关注该用户")
		}

		if err := tx.Model(&UserFollow{}).
			Where("follower_id = ? AND user_id = ?", userID, followerID).
			Update("status", 0).Error; err != nil {
			return xerr.Wrap(err, "downgrade reverse follow status failed")
		}

		return nil
	})
}

func CreateUserFollow(ctx context.Context, db *gorm.DB, followerID, userID string) error {
	if followerID == "" || userID == "" {
		return xerr.NewInvalidParam("关注者ID或用户ID为空")
	}

	relation := &UserFollow{
		FollowerID: followerID,
		UserID:     userID,
		Status:     0,
	}

	if err := db.WithContext(ctx).Create(relation).Error; err != nil {
		return xerr.Wrap(err, "create user follow failed")
	}

	return nil
}

func UpdateUserFollowStatus(ctx context.Context, db *gorm.DB, followerID, userID string, status int32) error {
	result := db.WithContext(ctx).
		Model(&UserFollow{}).
		Where("follower_id = ? AND user_id = ?", followerID, userID).
		Update("status", status)
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update user follow status failed")
	}

	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("用户关注关系不存在")
	}

	return nil
}

func GetFollowingISSubriber(ctx context.Context, db *gorm.DB, followerID, userID string) (bool, error) {
	var relation UserFollow
	err := db.WithContext(ctx).Where("follower_id = ? AND user_id = ?", followerID, userID).First(&relation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, xerr.Wrap(err, "get following status failed")
	}
	return true, nil
}

func GetFollowingByFollowerID(ctx context.Context, db *gorm.DB, followerID string, pageNumber, pageSize int32) ([]UserFollow, int64, error) {
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&UserFollow{}).Where("follower_id = ?", followerID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "count following list failed")
	}

	var relations []UserFollow
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&relations).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get following list failed")
	}

	return relations, total, nil
}

func GetFansByUserID(ctx context.Context, db *gorm.DB, userID string, pageNumber, pageSize int32) ([]UserFollow, int64, error) {
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&UserFollow{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "count fans list failed")
	}

	var relations []UserFollow
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&relations).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get fans list failed")
	}

	return relations, total, nil
}

func GetFriendByUserID(ctx context.Context, db *gorm.DB, userID string, pageNumber, pageSize int32) ([]UserFollow, int64, error) {
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&UserFollow{}).Where("user_id = ? AND status = ?", userID, 1)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "count friends list failed")
	}

	var relations []UserFollow
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&relations).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get friends list failed")
	}

	return relations, total, nil
}

package dal

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func CreateUser(ctx context.Context, user *types.UserBaseinfo) error {
	logger := logx.WithContext(ctx)

	if user == nil {
		err := errors.New("user is nil")
		logger.Errorf("create user failed: %v", err)
		return err
	}

	if err := Db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Errorf("create user failed: %v", err)
		return err
	}

	return nil
}

func GetUserByID(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
	logger := logx.WithContext(ctx)

	var user types.UserBaseinfo
	err := Db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("get user by id not found: %v", err)
			return nil, err
		}

		logger.Errorf("get user by id failed: %v", err)
		return nil, err
	}

	return &user, nil
}

func GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error) {
	logger := logx.WithContext(ctx)

	var user types.UserBaseinfo
	err := Db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("get user by username not found: %v", err)
			return nil, err
		}

		logger.Errorf("get user by username failed: %v", err)
		return nil, err
	}

	return &user, nil
}

func UpdateUserPhotoByID(ctx context.Context, userID string, photoURL string) error {
	logger := logx.WithContext(ctx)

	result := Db.WithContext(ctx).Model(&types.UserBaseinfo{}).Where("user_id = ?", userID).Update("photo_url", photoURL)
	if result.Error != nil {
		logger.Errorf("update user photo failed: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("update user photo failed, user not found: %v", err)
		return err
	}

	return nil
}

func GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
	logger := logx.WithContext(ctx)

	if len(userIDs) == 0 {
		return []types.UserBaseinfo{}, nil
	}

	var users []types.UserBaseinfo
	if err := Db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
		logger.Errorf("get users by ids failed: %v", err)
		return nil, err
	}

	return users, nil
}

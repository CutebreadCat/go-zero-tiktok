package user_baseinfo

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func CreateUser(ctx context.Context, db *gorm.DB, user *types.UserBaseinfo) error {
	logger := logx.WithContext(ctx)

	if user == nil {
		err := errors.New("user is nil")
		logger.Errorf("create user failed: %v", err)
		return xerr.New(400, "用户不存在")
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Errorf("create user failed: %v", err)
		return errors.New("创建用户失败")
	}

	return nil
}

func GetUserByID(ctx context.Context, db *gorm.DB, userID string) (*types.UserBaseinfo, error) {
	logger := logx.WithContext(ctx)

	var user types.UserBaseinfo
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("get user by id not found: %v", err)
			return nil, xerr.New(400, "用户不存在")
		}

		logger.Errorf("get user by id failed: %v", err)
		return nil, errors.New("获取用户失败")
	}

	return &user, nil
}

func GetUserByUsername(ctx context.Context, db *gorm.DB, username string) (*types.UserBaseinfo, error) {
	logger := logx.WithContext(ctx)

	var user types.UserBaseinfo
	err := db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("get user by username not found: %v", err)
			return nil, xerr.New(400, "用户不存在")
		}

		logger.Errorf("get user by username failed: %v", err)
		return nil, errors.New("获取用户失败")
	}

	return &user, nil
}

func UpdateUserPhotoByID(ctx context.Context, db *gorm.DB, userID string, photoURL string) error {
	logger := logx.WithContext(ctx)

	result := db.WithContext(ctx).Model(&types.UserBaseinfo{}).Where("user_id = ?", userID).Update("photo_url", photoURL)
	if result.Error != nil {
		logger.Errorf("update user photo failed: %v", result.Error)
		return errors.New("更新用户头像失败")
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("update user photo failed, user not found: %v", err)
		return xerr.New(400, "没有进行更新")
	}

	return nil
}

func GetUsersByIDs(ctx context.Context, db *gorm.DB, userIDs []string) ([]types.UserBaseinfo, error) {
	logger := logx.WithContext(ctx)

	if len(userIDs) == 0 {
		return []types.UserBaseinfo{}, nil
	}

	var users []types.UserBaseinfo
	if err := db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
		logger.Errorf("get users by ids failed: %v", err)
		return nil, errors.New("获取用户列表失败")
	}

	return users, nil
}

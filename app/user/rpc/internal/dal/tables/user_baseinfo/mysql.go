package user_baseinfo

import (
	"context"
	"errors"

	"go_zero-tiktok/pkg/xerr"

	"gorm.io/gorm"
)

func CreateUser(ctx context.Context, db *gorm.DB, user *UserBaseinfo) error {
	if user == nil {
		return xerr.NewInvalidParam("用户不存在")
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		return xerr.Wrap(err, "create user failed")
	}

	return nil
}

func GetUserByID(ctx context.Context, db *gorm.DB, userID int64) (*UserBaseinfo, error) {
	var user UserBaseinfo
	err := db.WithContext(ctx).Where("user_id = ? AND status = 1", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerr.NewInvalidParam("用户不存在")
		}
		return nil, xerr.Wrap(err, "get user by id failed")
	}

	return &user, nil
}

func GetUserByUsername(ctx context.Context, db *gorm.DB, username string) (*UserBaseinfo, error) {
	var user UserBaseinfo
	err := db.WithContext(ctx).Where("username = ? AND status = 1", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerr.NewInvalidParam("用户不存在")
		}
		return nil, xerr.Wrap(err, "get user by username failed")
	}
	return &user, nil
}

func UpdateUserPhotoByID(ctx context.Context, db *gorm.DB, userID int64, photoURL string) error {
	result := db.WithContext(ctx).Model(&UserBaseinfo{}).Where("user_id = ? AND status = 1", userID).Update("photo_url", photoURL)
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update user photo failed")
	}
	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("没有进行更新")
	}

	return nil
}

func GetUsersByIDs(ctx context.Context, db *gorm.DB, userIDs []int64) ([]UserBaseinfo, error) {
	if len(userIDs) == 0 {
		return []UserBaseinfo{}, nil
	}

	var users []UserBaseinfo
	if err := db.WithContext(ctx).Where("user_id IN ? AND status = 1", userIDs).Find(&users).Error; err != nil {
		return nil, xerr.Wrap(err, "get users by ids failed")
	}
	return users, nil
}

// DeleteUserByID 软删除用户：status 置 0（0已删），同一用户名可重新注册。
func DeleteUserByID(ctx context.Context, db *gorm.DB, userID int64) error {
	result := db.WithContext(ctx).Model(&UserBaseinfo{}).
		Where("user_id = ? AND status = 1", userID).
		Update("status", 0)
	if result.Error != nil {
		return xerr.Wrap(result.Error, "delete user failed")
	}
	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("用户不存在或已删除")
	}
	return nil
}

func CheckUserExistsMFA(ctx context.Context, db *gorm.DB, userID int64) (bool, error) {
	var user UserBaseinfo
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, xerr.Wrap(err, "check user mfa failed")
	}
	return user.MFAEnabled, nil
}

func UpdateUserMFAPendingSecret(ctx context.Context, db *gorm.DB, userID int64, secret string) error {
	result := db.WithContext(ctx).Model(&UserBaseinfo{}).Where("user_id = ?", userID).Update("mfa_pending_secret", secret)
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update user mfa secret failed")
	}
	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("没有进行更新")
	}
	return nil
}

// FindUserMFASecret 查询用户已启用的 MFA 正式 secret。
func FindUserMFASecret(ctx context.Context, db *gorm.DB, userID int64) (string, error) {
	var user UserBaseinfo
	err := db.WithContext(ctx).Model(&UserBaseinfo{}).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", xerr.NewInvalidParam("用户MFA信息不存在")
		}
		return "", xerr.Wrap(err, "find user mfa secret failed")
	}
	return user.MFASecret, nil
}

// FindUserPendMFASecret 查询用户待绑定（pending）的 MFA secret。
func FindUserPendMFASecret(ctx context.Context, db *gorm.DB, userID int64) (string, error) {
	var user UserBaseinfo
	err := db.WithContext(ctx).Model(&UserBaseinfo{}).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", xerr.NewInvalidParam("用户MFA信息不存在")
		}
		return "", xerr.Wrap(err, "find user mfa secret failed")
	}
	return user.MFAPendingSecret, nil
}

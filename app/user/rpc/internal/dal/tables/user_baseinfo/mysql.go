package user_baseinfo

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/shared/xerr"

	"gorm.io/gorm"
)

func CreateUser(ctx context.Context, db *gorm.DB, user *UserBaseinfo) error {
	if user == nil {
		return xerr.NewInvalidParam("用户不存在")
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		return xerr.Wrap(err, "create user failed")
	}
	Usermaf := &UserMFA{
		UserID:           user.UserID,
		PasswordHash:     user.Password,
		MFASecret:        "",
		MFAPendingSecret: "",
	}
	if err := db.WithContext(ctx).Create(Usermaf).Error; err != nil {
		return xerr.Wrap(err, "create user mfa failed")
	}

	return nil
}

func GetUserByID(ctx context.Context, db *gorm.DB, userID string) (*UserBaseinfo, error) {
	var user UserBaseinfo
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
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
	err := db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerr.NewInvalidParam("用户不存在")
		}
		return nil, xerr.Wrap(err, "get user by username failed")
	}
	return &user, nil
}

func UpdateUserPhotoByID(ctx context.Context, db *gorm.DB, userID string, photoURL string) error {
	result := db.WithContext(ctx).Model(&UserBaseinfo{}).Where("user_id = ?", userID).Update("photo_url", photoURL)
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update user photo failed")
	}
	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("没有进行更新")
	}

	return nil
}

func GetUsersByIDs(ctx context.Context, db *gorm.DB, userIDs []string) ([]UserBaseinfo, error) {
	if len(userIDs) == 0 {
		return []UserBaseinfo{}, nil
	}

	var users []UserBaseinfo
	if err := db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, xerr.Wrap(err, "get users by ids failed")
	}
	return users, nil
}

func CheckUserExistsMFA(ctx context.Context, db *gorm.DB, userID string) (bool, error) {
	var userMFA UserMFA
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&userMFA).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, xerr.Wrap(err, "check user mfa failed")
	}
	return userMFA.MFAEnabled, nil
}

func UpdateUserMFAPendingSecret(ctx context.Context, db *gorm.DB, userID string, secret string) error {
	result := db.WithContext(ctx).Model(&UserMFA{}).Where("user_id = ?", userID).Update("mfa_pending_secret", secret)
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update user mfa secret failed")
	}
	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("没有进行更新")
	}
	return nil
}

func FindUserMFASecret(ctx context.Context, db *gorm.DB, userID string) (string, error) {
	var userMFA UserMFA
	err := db.WithContext(ctx).Model(&UserMFA{}).Where("user_id = ?", userID).First(&userMFA).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", xerr.NewInvalidParam("用户MFA信息不存在")
		}
		return "", xerr.Wrap(err, "find user mfa secret failed")
	}
	return userMFA.MFAPendingSecret, nil
}

func FindUserPendMFASecret(ctx context.Context, db *gorm.DB, userID string) (string, error) {
	var userMFA UserMFA
	err := db.WithContext(ctx).Model(&UserMFA{}).Where("user_id = ?", userID).First(&userMFA).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", xerr.NewInvalidParam("用户MFA信息不存在")
		}
		return "", xerr.Wrap(err, "find user mfa secret failed")
	}
	return userMFA.MFASecret, nil
}

func UpdateUserJwchInfo(ctx context.Context, db *gorm.DB, userID string, jwchID string, jwchPassword string) error {
	result := db.WithContext(ctx).Model(&UserBaseinfo{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"jwch_id":       jwchID,
		"jwch_password": jwchPassword,
	})
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update user jwch info failed")
	}
	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("没有进行更新")
	}
	return nil
}

func GetUserJwchInfo(ctx context.Context, db *gorm.DB, userID string) (string, string, error) {
	var user UserBaseinfo
	err := db.WithContext(ctx).Model(&UserBaseinfo{}).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", xerr.NewInvalidParam("用户不存在")
		}
		return "", "", xerr.Wrap(err, "get user jwch info failed")
	}
	return user.JwchID, user.JwchPassword, nil
}

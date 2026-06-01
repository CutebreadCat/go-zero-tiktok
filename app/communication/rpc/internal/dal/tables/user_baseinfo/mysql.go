package user_baseinfo

import (
	"context"

	"go_zero-tiktok/internal/shared/xerr"

	"gorm.io/gorm"
)

func GetUsersByIDs(ctx context.Context, db *gorm.DB, userIDs []string) ([]UserBaseinfo, error) {
	var users []UserBaseinfo
	if err := db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, xerr.Wrap(err, "get users by ids failed")
	}
	return users, nil
}

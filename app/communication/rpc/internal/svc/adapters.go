package svc

import (
	"context"

	userbasetable "go_zero-tiktok/app/communication/rpc/internal/dal/tables/user_baseinfo"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"gorm.io/gorm"
)

// UserRepoAdapter 用户仓储适配器
type UserRepoAdapter struct {
	db *gorm.DB
}

func NewUserRepoAdapter(db *gorm.DB) *UserRepoAdapter {
	return &UserRepoAdapter{db: db}
}

func (a *UserRepoAdapter) GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
	users, err := userbasetable.GetUsersByIDs(ctx, a.db, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]types.UserBaseinfo, 0, len(users))
	for _, u := range users {
		result = append(result, types.UserBaseinfo{
			UserID:    u.UserID,
			Username:  u.Username,
			PhotoURL:  u.PhotoURL,
			CreatedAt: myutils.TimeToStr(u.CreatedAt, ""),
			UpdatedAt: myutils.TimeToStr(u.UpdatedAt, ""),
		})
	}
	return result, nil
}

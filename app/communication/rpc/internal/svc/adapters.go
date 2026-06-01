package svc

import (
	"context"

	userrepo "go_zero-tiktok/internal/dal/repository"
	"go_zero-tiktok/internal/types"

	"gorm.io/gorm"
)

// UserRepoAdapter 用户仓储适配器（直接使用现有 internal 的 user repo）
type UserRepoAdapter struct {
	repo *userrepo.UserBaseinfoRepo
}

func NewUserRepoAdapter(db *gorm.DB) *UserRepoAdapter {
	return &UserRepoAdapter{
		repo: userrepo.NewUserBaseinfoRepo(db),
	}
}

func (a *UserRepoAdapter) GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
	return a.repo.GetUsersByIDs(ctx, userIDs)
}

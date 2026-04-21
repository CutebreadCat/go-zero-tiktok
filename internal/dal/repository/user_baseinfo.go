package repository

import (
	"context"

	userbasetable "go_zero-tiktok/internal/dal/tables/user_baseinfo"
	"go_zero-tiktok/internal/types"

	"gorm.io/gorm"
)

type UserBaseinfoRepo struct {
	db *gorm.DB
}

func NewUserBaseinfoRepo(db *gorm.DB) *UserBaseinfoRepo {
	return &UserBaseinfoRepo{db: db}
}

func (r *UserBaseinfoRepo) CreateUser(ctx context.Context, user *types.UserBaseinfo) error {
	return userbasetable.CreateUser(ctx, r.db, user)
}

func (r *UserBaseinfoRepo) GetUserByID(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
	return userbasetable.GetUserByID(ctx, r.db, userID)
}

func (r *UserBaseinfoRepo) GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error) {
	return userbasetable.GetUserByUsername(ctx, r.db, username)
}

func (r *UserBaseinfoRepo) UpdateUserPhotoByID(ctx context.Context, userID string, photoURL string) error {
	return userbasetable.UpdateUserPhotoByID(ctx, r.db, userID, photoURL)
}

func (r *UserBaseinfoRepo) GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
	return userbasetable.GetUsersByIDs(ctx, r.db, userIDs)
}

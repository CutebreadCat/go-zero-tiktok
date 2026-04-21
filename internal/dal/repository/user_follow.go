package repository

import (
	"context"

	userfollowtable "go_zero-tiktok/internal/dal/tables/user_follow"
	"go_zero-tiktok/internal/types"

	"gorm.io/gorm"
)

type UserFollowRepo struct {
	db *gorm.DB
}

func NewUserFollowRepo(db *gorm.DB) *UserFollowRepo {
	return &UserFollowRepo{db: db}
}

func (r *UserFollowRepo) CreateUserFollow(ctx context.Context, followerID, userID string) error {
	return userfollowtable.CreateUserFollow(ctx, r.db, followerID, userID)
}

func (r *UserFollowRepo) FollowUser(ctx context.Context, followerID, userID string) error {
	return userfollowtable.FollowUser(ctx, r.db, followerID, userID)
}

func (r *UserFollowRepo) UnfollowUser(ctx context.Context, followerID, userID string) error {
	return userfollowtable.UnfollowUser(ctx, r.db, followerID, userID)
}

func (r *UserFollowRepo) UpdateUserFollowStatus(ctx context.Context, followerID, userID string, status int32) error {
	return userfollowtable.UpdateUserFollowStatus(ctx, r.db, followerID, userID, status)
}

func (r *UserFollowRepo) GetFollowingISSubriber(ctx context.Context, followerID, userID string) (bool, error) {
	return userfollowtable.GetFollowingISSubriber(ctx, r.db, followerID, userID)
}

func (r *UserFollowRepo) GetFollowingByFollowerID(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return userfollowtable.GetFollowingByFollowerID(ctx, r.db, followerID, pageNumber, pageSize)
}

func (r *UserFollowRepo) GetFansByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return userfollowtable.GetFansByUserID(ctx, r.db, userID, pageNumber, pageSize)
}

func (r *UserFollowRepo) GetFriendByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return userfollowtable.GetFriendByUserID(ctx, r.db, userID, pageNumber, pageSize)
}

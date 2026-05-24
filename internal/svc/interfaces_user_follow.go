package svc

import (
	"context"

	"go_zero-tiktok/internal/types"
)

type IUserFollowRepo interface {
	FollowUser(ctx context.Context, followerID, userID string) error
	UnfollowUser(ctx context.Context, followerID, userID string) error
	GetFollowingByFollowerID(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
	GetFansByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
	GetFriendByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
}

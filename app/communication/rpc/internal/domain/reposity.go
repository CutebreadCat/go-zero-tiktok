package domain

import (
	"context"
	"go_zero-tiktok/pkg/contract"
)

type IUserFollowRepo interface {
	FollowUser(ctx context.Context, followerID, userID int64) error
	UnfollowUser(ctx context.Context, followerID, userID int64) error
	GetFollowingByFollowerID(ctx context.Context, followerID int64, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
	GetFansByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
	GetFriendByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
}

type IUserRepo interface {
	GetUsersByIDs(ctx context.Context, userIDs []int64) ([]types.UserBaseinfo, error)
}

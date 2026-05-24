package mock

import (
	"context"

	"go_zero-tiktok/internal/types"
)

type UserFollowRepo struct {
	FollowUserFn               func(ctx context.Context, followerID, userID string) error
	UnfollowUserFn             func(ctx context.Context, followerID, userID string) error
	GetFollowingByFollowerIDFn func(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
	GetFansByUserIDFn          func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
	GetFriendByUserIDFn        func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
}

func (m *UserFollowRepo) FollowUser(ctx context.Context, followerID, userID string) error {
	if m.FollowUserFn != nil {
		return m.FollowUserFn(ctx, followerID, userID)
	}
	return nil
}

func (m *UserFollowRepo) UnfollowUser(ctx context.Context, followerID, userID string) error {
	if m.UnfollowUserFn != nil {
		return m.UnfollowUserFn(ctx, followerID, userID)
	}
	return nil
}

func (m *UserFollowRepo) GetFollowingByFollowerID(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	if m.GetFollowingByFollowerIDFn != nil {
		return m.GetFollowingByFollowerIDFn(ctx, followerID, pageNumber, pageSize)
	}
	return nil, 0, nil
}

func (m *UserFollowRepo) GetFansByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	if m.GetFansByUserIDFn != nil {
		return m.GetFansByUserIDFn(ctx, userID, pageNumber, pageSize)
	}
	return nil, 0, nil
}

func (m *UserFollowRepo) GetFriendByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	if m.GetFriendByUserIDFn != nil {
		return m.GetFriendByUserIDFn(ctx, userID, pageNumber, pageSize)
	}
	return nil, 0, nil
}

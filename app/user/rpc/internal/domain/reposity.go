package user_service

import (
	"context"
	"go_zero-tiktok/pkg/contract"
)

type IUserRepo interface {
	CreateUserFromParams(ctx context.Context, userID, username, password, photoURL string) error
	GetUserByID(ctx context.Context, userID string) (*types.UserBaseinfo, error)
	GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error)
	GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error)
	UpdateUserPhotoByID(ctx context.Context, userID string, photoURL string) error
	CheckExistsMFA(ctx context.Context, userID string) (bool, error)
	UpdateUserMFAPendingSecret(ctx context.Context, userID string, pendingSecret string) error
	EnableUserMFA(ctx context.Context, userID string) error
	FindUserMFASecret(ctx context.Context, userID string) (string, error)
	FindUserPendMFASecret(ctx context.Context, userID string) (string, error)
	UpdateUserJwchInfo(ctx context.Context, userID string, jwchID string, jwchPassword string) error
	GetUserJwchInfo(ctx context.Context, userID string) (string, string, error)
}

type IUserFollowRepo interface {
	FollowUser(ctx context.Context, followerID, userID string) error
	UnfollowUser(ctx context.Context, followerID, userID string) error
	GetFollowingByFollowerID(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
	GetFansByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
	GetFriendByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error)
}

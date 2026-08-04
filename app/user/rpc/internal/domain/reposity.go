package user_service

import (
	"context"
	"go_zero-tiktok/pkg/contract"
)

type IUserRepo interface {
	CreateUserFromParams(ctx context.Context, userID int64, username, password, photoURL string) error
	GetUserByID(ctx context.Context, userID int64) (*types.UserBaseinfo, error)
	GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error)
	GetUsersByIDs(ctx context.Context, userIDs []int64) ([]types.UserBaseinfo, error)
	UpdateUserPhotoByID(ctx context.Context, userID int64, photoURL string) error
	DeleteUserByID(ctx context.Context, userID int64) error
	CheckExistsMFA(ctx context.Context, userID int64) (bool, error)
	UpdateUserMFAPendingSecret(ctx context.Context, userID int64, pendingSecret string) error
	EnableUserMFA(ctx context.Context, userID int64) error
	FindUserMFASecret(ctx context.Context, userID int64) (string, error)
	FindUserPendMFASecret(ctx context.Context, userID int64) (string, error)
}

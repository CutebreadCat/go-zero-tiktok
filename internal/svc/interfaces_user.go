package svc

import (
	"context"

	"go_zero-tiktok/internal/types"
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

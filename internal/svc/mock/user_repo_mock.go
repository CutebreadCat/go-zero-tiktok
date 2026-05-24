package mock

import (
	"context"

	"go_zero-tiktok/internal/types"
)

type UserRepo struct {
	CreateUserFromParamsFn   func(ctx context.Context, userID, username, password, photoURL string) error
	GetUserByIDFn            func(ctx context.Context, userID string) (*types.UserBaseinfo, error)
	GetUserByUsernameFn      func(ctx context.Context, username string) (*types.UserBaseinfo, error)
	GetUsersByIDsFn          func(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error)
	UpdateUserPhotoByIDFn    func(ctx context.Context, userID string, photoURL string) error
	CheckExistsMFAFn         func(ctx context.Context, userID string) (bool, error)
	UpdateUserMFAPendingFn   func(ctx context.Context, userID string, pendingSecret string) error
	EnableUserMFAFn          func(ctx context.Context, userID string) error
	FindUserMFASecretFn      func(ctx context.Context, userID string) (string, error)
	FindUserPendMFASecretFn  func(ctx context.Context, userID string) (string, error)
	UpdateUserJwchInfoFn     func(ctx context.Context, userID string, jwchID string, jwchPassword string) error
	GetUserJwchInfoFn        func(ctx context.Context, userID string) (string, string, error)
}

func (m *UserRepo) CreateUserFromParams(ctx context.Context, userID, username, password, photoURL string) error {
	if m.CreateUserFromParamsFn != nil {
		return m.CreateUserFromParamsFn(ctx, userID, username, password, photoURL)
	}
	return nil
}

func (m *UserRepo) GetUserByID(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *UserRepo) GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error) {
	if m.GetUserByUsernameFn != nil {
		return m.GetUserByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *UserRepo) GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
	if m.GetUsersByIDsFn != nil {
		return m.GetUsersByIDsFn(ctx, userIDs)
	}
	return nil, nil
}

func (m *UserRepo) UpdateUserPhotoByID(ctx context.Context, userID string, photoURL string) error {
	if m.UpdateUserPhotoByIDFn != nil {
		return m.UpdateUserPhotoByIDFn(ctx, userID, photoURL)
	}
	return nil
}

func (m *UserRepo) CheckExistsMFA(ctx context.Context, userID string) (bool, error) {
	if m.CheckExistsMFAFn != nil {
		return m.CheckExistsMFAFn(ctx, userID)
	}
	return false, nil
}

func (m *UserRepo) UpdateUserMFAPendingSecret(ctx context.Context, userID string, pendingSecret string) error {
	if m.UpdateUserMFAPendingFn != nil {
		return m.UpdateUserMFAPendingFn(ctx, userID, pendingSecret)
	}
	return nil
}

func (m *UserRepo) EnableUserMFA(ctx context.Context, userID string) error {
	if m.EnableUserMFAFn != nil {
		return m.EnableUserMFAFn(ctx, userID)
	}
	return nil
}

func (m *UserRepo) FindUserMFASecret(ctx context.Context, userID string) (string, error) {
	if m.FindUserMFASecretFn != nil {
		return m.FindUserMFASecretFn(ctx, userID)
	}
	return "", nil
}

func (m *UserRepo) FindUserPendMFASecret(ctx context.Context, userID string) (string, error) {
	if m.FindUserPendMFASecretFn != nil {
		return m.FindUserPendMFASecretFn(ctx, userID)
	}
	return "", nil
}

func (m *UserRepo) UpdateUserJwchInfo(ctx context.Context, userID string, jwchID string, jwchPassword string) error {
	if m.UpdateUserJwchInfoFn != nil {
		return m.UpdateUserJwchInfoFn(ctx, userID, jwchID, jwchPassword)
	}
	return nil
}

func (m *UserRepo) GetUserJwchInfo(ctx context.Context, userID string) (string, string, error) {
	if m.GetUserJwchInfoFn != nil {
		return m.GetUserJwchInfoFn(ctx, userID)
	}
	return "", "", nil
}

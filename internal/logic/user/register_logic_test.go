package user

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestRegisterLogic_Register(t *testing.T) {
	tests := []struct {
		name     string
		req      *types.RegisterRequest
		userRepo *mock.UserRepo
		wantErr  bool
	}{
		{
			name:    "empty username",
			req:     &types.RegisterRequest{Username: "", Password: "123456"},
			wantErr: true,
		},
		{
			name:    "empty password",
			req:     &types.RegisterRequest{Username: "test", Password: ""},
			wantErr: true,
		},
		{
			name: "username already exists",
			req:  &types.RegisterRequest{Username: "existing", Password: "123456"},
			userRepo: &mock.UserRepo{
				GetUserByUsernameFn: func(ctx context.Context, username string) (*types.UserBaseinfo, error) {
					return &types.UserBaseinfo{UserID: "u1", Username: "existing"}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "create user failed",
			req:  &types.RegisterRequest{Username: "newuser", Password: "123456"},
			userRepo: &mock.UserRepo{
				GetUserByUsernameFn: func(ctx context.Context, username string) (*types.UserBaseinfo, error) {
					return nil, errors.New("not found")
				},
				CreateUserFromParamsFn: func(ctx context.Context, userID, username, password, photoURL string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			req:  &types.RegisterRequest{Username: "newuser", Password: "123456"},
			userRepo: &mock.UserRepo{
				GetUserByUsernameFn: func(ctx context.Context, username string) (*types.UserBaseinfo, error) {
					return nil, errors.New("not found")
				},
				CreateUserFromParamsFn: func(ctx context.Context, userID, username, password, photoURL string) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(tt.userRepo, nil, nil, nil, nil, nil, nil)
			logic := NewRegisterLogic(context.Background(), svcCtx)
			resp, err := logic.Register(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if resp.UserID == "" {
				t.Errorf("expected userID, got empty")
			}
		})
	}
}

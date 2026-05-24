package user

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"
)

func TestLoginLogic_Login(t *testing.T) {
	hashedPwd := myutils.HashPassword("123456")

	tests := []struct {
		name     string
		req      *types.LoginRequest
		userRepo *mock.UserRepo
		wantErr  bool
	}{
		{
			name:    "empty username",
			req:     &types.LoginRequest{Username: "", Password: "123456"},
			wantErr: true,
		},
		{
			name:    "empty password",
			req:     &types.LoginRequest{Username: "test", Password: ""},
			wantErr: true,
		},
		{
			name: "user not found",
			req:  &types.LoginRequest{Username: "notexist", Password: "123456"},
			userRepo: &mock.UserRepo{
				GetUserByUsernameFn: func(ctx context.Context, username string) (*types.UserBaseinfo, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
		{
			name: "wrong password",
			req:  &types.LoginRequest{Username: "test", Password: "wrong"},
			userRepo: &mock.UserRepo{
				GetUserByUsernameFn: func(ctx context.Context, username string) (*types.UserBaseinfo, error) {
					return &types.UserBaseinfo{UserID: "u1", Username: "test", Password: hashedPwd}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "mfa required but code empty",
			req:  &types.LoginRequest{Username: "test", Password: "123456", MfaCode: ""},
			userRepo: &mock.UserRepo{
				GetUserByUsernameFn: func(ctx context.Context, username string) (*types.UserBaseinfo, error) {
					return &types.UserBaseinfo{UserID: "u1", Username: "test", Password: hashedPwd}, nil
				},
				CheckExistsMFAFn: func(ctx context.Context, userID string) (bool, error) {
					return true, nil
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(tt.userRepo, nil, nil, nil, nil, nil, nil)
			logic := NewLoginLogic(context.Background(), svcCtx)
			resp, err := logic.Login(tt.req)
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
			if resp.UserID == "" || resp.AccessToken == "" || resp.RefreshToken == "" {
				t.Errorf("expected non-empty response fields")
			}
		})
	}
}

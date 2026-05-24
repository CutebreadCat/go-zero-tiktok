package user

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestGetUserInfoLogic_GetUserInfo(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		req      *types.UserInfoRequest
		userRepo *mock.UserRepo
		wantErr  bool
	}{
		{
			name:    "empty user id",
			ctx:     context.Background(),
			req:     &types.UserInfoRequest{UserID: ""},
			wantErr: true,
		},
		{
			name: "user not found",
			ctx:  context.Background(),
			req:  &types.UserInfoRequest{UserID: "u1"},
			userRepo: &mock.UserRepo{
				GetUserByIDFn: func(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  context.Background(),
			req:  &types.UserInfoRequest{UserID: "u1"},
			userRepo: &mock.UserRepo{
				GetUserByIDFn: func(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
					return &types.UserBaseinfo{UserID: "u1", Username: "test"}, nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(tt.userRepo, nil, nil, nil, nil, nil, nil)
			logic := NewGetUserInfoLogic(tt.ctx, svcCtx)
			resp, err := logic.GetUserInfo(tt.req)
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
			if resp.User.UserID == "" {
				t.Errorf("expected user ID, got empty")
			}
		})
	}
}

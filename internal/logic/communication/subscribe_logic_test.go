package communication

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestSubscribeLogic_Subscribe(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		req        *types.SubscribeRequest
		userFollow *mock.UserFollowRepo
		wantErr    bool
	}{
		{
			name:    "empty to user id",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.SubscribeRequest{ToUserID: "", ActionType: 1},
			wantErr: true,
		},
		{
			name:    "invalid action type",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.SubscribeRequest{ToUserID: "u2", ActionType: 2},
			wantErr: true,
		},
		{
			name: "follow failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.SubscribeRequest{ToUserID: "u2", ActionType: 1},
			userFollow: &mock.UserFollowRepo{
				FollowUserFn: func(ctx context.Context, followerID, userID string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "unfollow failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.SubscribeRequest{ToUserID: "u2", ActionType: 0},
			userFollow: &mock.UserFollowRepo{
				UnfollowUserFn: func(ctx context.Context, followerID, userID string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success follow",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.SubscribeRequest{ToUserID: "u2", ActionType: 1},
			userFollow: &mock.UserFollowRepo{
				FollowUserFn: func(ctx context.Context, followerID, userID string) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "success unfollow",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.SubscribeRequest{ToUserID: "u2", ActionType: 0},
			userFollow: &mock.UserFollowRepo{
				UnfollowUserFn: func(ctx context.Context, followerID, userID string) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, nil, nil, tt.userFollow, nil)
			logic := NewSubscribeLogic(tt.ctx, svcCtx)
			_, err := logic.Subscribe(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

package communication

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestGetFriendListLogic_GetFriendList(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		req        *types.GetFriendListRequest
		userFollow *mock.UserFollowRepo
		userRepo   *mock.UserRepo
		wantErr    bool
		wantLen    int
	}{
		{
			name: "get friends failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.GetFriendListRequest{},
			userFollow: &mock.UserFollowRepo{
				GetFriendByUserIDFn: func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "get users failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.GetFriendListRequest{},
			userFollow: &mock.UserFollowRepo{
				GetFriendByUserIDFn: func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
					return []types.UserFollow{{FollowerID: "u2", UserID: "u1"}}, 1, nil
				},
			},
			userRepo: &mock.UserRepo{
				GetUsersByIDsFn: func(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.GetFriendListRequest{PageNumber: 1, PageSize: 10},
			userFollow: &mock.UserFollowRepo{
				GetFriendByUserIDFn: func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
					return []types.UserFollow{
						{FollowerID: "u2", UserID: "u1"},
						{FollowerID: "u3", UserID: "u1"},
					}, 2, nil
				},
			},
			userRepo: &mock.UserRepo{
				GetUsersByIDsFn: func(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
					return []types.UserBaseinfo{
						{UserID: "u2", Username: "user2"},
						{UserID: "u3", Username: "user3"},
					}, nil
				},
			},
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(tt.userRepo, nil, nil, nil, nil, tt.userFollow, nil)
			logic := NewGetFriendListLogic(tt.ctx, svcCtx)
			resp, err := logic.GetFriendList(tt.req)
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
			if len(resp.FriendList) != tt.wantLen {
				t.Errorf("expected %d friends, got %d", tt.wantLen, len(resp.FriendList))
			}
		})
	}
}

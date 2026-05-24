package communication

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestGetSubscriberListLogic_GetSubscriberList(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		req        *types.GetSubscriberListRequest
		userFollow *mock.UserFollowRepo
		userRepo   *mock.UserRepo
		wantErr    bool
		wantLen    int
	}{
		{
			name: "get following failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.GetSubscriberListRequest{},
			userFollow: &mock.UserFollowRepo{
				GetFollowingByFollowerIDFn: func(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.GetSubscriberListRequest{PageNumber: 1, PageSize: 10},
			userFollow: &mock.UserFollowRepo{
				GetFollowingByFollowerIDFn: func(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
					return []types.UserFollow{
						{FollowerID: "u1", UserID: "u2"},
						{FollowerID: "u1", UserID: "u3"},
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
			logic := NewGetSubscriberListLogic(tt.ctx, svcCtx)
			resp, err := logic.GetSubscriberList(tt.req)
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
			if len(resp.SubscriberList) != tt.wantLen {
				t.Errorf("expected %d subscribers, got %d", tt.wantLen, len(resp.SubscriberList))
			}
		})
	}
}

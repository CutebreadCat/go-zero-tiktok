package interaction

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestLikeVideoLogic_LikeVideo(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		req           *types.LikeVideoRequest
		videoLiker    *mock.VideoLikerRepo
		popular       *mock.PopularRepo
		wantErr       bool
	}{
		{
			name:    "empty video id",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.LikeVideoRequest{VideoID: "", ActionType: 1},
			wantErr: true,
		},
		{
			name:    "invalid action type",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.LikeVideoRequest{VideoID: "v1", ActionType: 2},
			wantErr: true,
		},
		{
			name: "like video failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.LikeVideoRequest{VideoID: "v1", ActionType: 1},
			videoLiker: &mock.VideoLikerRepo{
				LikeVideoFn: func(ctx context.Context, userID, videoID string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "cancel like video failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.LikeVideoRequest{VideoID: "v1", ActionType: 0},
			videoLiker: &mock.VideoLikerRepo{
				CancelLikeVideoFn: func(ctx context.Context, userID, videoID string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success like",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.LikeVideoRequest{VideoID: "v1", ActionType: 1},
			videoLiker: &mock.VideoLikerRepo{
				LikeVideoFn: func(ctx context.Context, userID, videoID string) error {
					return nil
				},
			},
			popular: &mock.PopularRepo{
				UpdateVideoLikeCountFn: func(ctx context.Context, videoID string, delta int64) error {
					return nil
				},
				IncreaseVideoVisitCountFn: func(ctx context.Context, videoID string, delta int64) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "success cancel like",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.LikeVideoRequest{VideoID: "v1", ActionType: 0},
			videoLiker: &mock.VideoLikerRepo{
				CancelLikeVideoFn: func(ctx context.Context, userID, videoID string) error {
					return nil
				},
			},
			popular: &mock.PopularRepo{
				UpdateVideoLikeCountFn: func(ctx context.Context, videoID string, delta int64) error {
					return nil
				},
				IncreaseVideoVisitCountFn: func(ctx context.Context, videoID string, delta int64) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, tt.popular, nil, tt.videoLiker, nil, nil)
			logic := NewLikeVideoLogic(tt.ctx, svcCtx)
			_, err := logic.LikeVideo(tt.req)
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

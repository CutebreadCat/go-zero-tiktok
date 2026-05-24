package interaction

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestGetLikeListLogic_GetLikeList(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		req        *types.GetLikeListRequest
		videoLiker *mock.VideoLikerRepo
		video      *mock.VideoRepo
		wantErr    bool
		wantLen    int
	}{
		{
			name: "get liked videos failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.GetLikeListRequest{},
			videoLiker: &mock.VideoLikerRepo{
				GetLikedVideoIDsByUserIDFn: func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "get videos by ids failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.GetLikeListRequest{},
			videoLiker: &mock.VideoLikerRepo{
				GetLikedVideoIDsByUserIDFn: func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
					return []string{"v1"}, 1, nil
				},
			},
			video: &mock.VideoRepo{
				GetVideosByIDsFn: func(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.GetLikeListRequest{},
			videoLiker: &mock.VideoLikerRepo{
				GetLikedVideoIDsByUserIDFn: func(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
					return []string{"v1", "v2"}, 2, nil
				},
			},
			video: &mock.VideoRepo{
				GetVideosByIDsFn: func(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error) {
					return []types.VideoBaseinfo{
						{VideoID: "v1", AuthorID: "u2"},
						{VideoID: "v2", AuthorID: "u3"},
					}, nil
				},
			},
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, tt.video, nil, nil, tt.videoLiker, nil, nil)
			logic := NewGetLikeListLogic(tt.ctx, svcCtx)
			resp, err := logic.GetLikeList(tt.req)
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
			if len(resp.VideoList) != tt.wantLen {
				t.Errorf("expected %d videos, got %d", tt.wantLen, len(resp.VideoList))
			}
		})
	}
}

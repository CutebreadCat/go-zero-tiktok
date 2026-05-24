package video

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestVideoPopularLogic_VideoPopular(t *testing.T) {
	tests := []struct {
		name    string
		req     *types.VideoPopularRequest
		video   *mock.VideoRepo
		popular *mock.PopularRepo
		wantErr bool
		wantLen int
	}{
		{
			name: "get popular failed",
			req:  &types.VideoPopularRequest{},
			popular: &mock.PopularRepo{
				GetPopularVideoIDsByVisitCountFn: func(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error) {
					return nil, 0, errors.New("redis error")
				},
			},
			wantErr: true,
		},
		{
			name: "get videos failed",
			req:  &types.VideoPopularRequest{},
			popular: &mock.PopularRepo{
				GetPopularVideoIDsByVisitCountFn: func(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error) {
					return []types.VideoPopular{{VideoID: "v1"}}, 1, nil
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
			req:  &types.VideoPopularRequest{PageSize: 10, PageNum: 1},
			popular: &mock.PopularRepo{
				GetPopularVideoIDsByVisitCountFn: func(ctx context.Context, pageNum, pageSize int32) ([]types.VideoPopular, int64, error) {
					return []types.VideoPopular{
						{VideoID: "v1", VisitCount: 100},
						{VideoID: "v2", VisitCount: 50},
					}, 2, nil
				},
			},
			video: &mock.VideoRepo{
				GetVideosByIDsFn: func(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error) {
					return []types.VideoBaseinfo{
						{VideoID: "v1", Title: "popular1"},
						{VideoID: "v2", Title: "popular2"},
					}, nil
				},
			},
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, tt.video, tt.popular, nil, nil, nil, nil)
			logic := NewVideoPopularLogic(context.Background(), svcCtx)
			resp, err := logic.VideoPopular(tt.req)
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
			if len(resp.Videos) != tt.wantLen {
				t.Errorf("expected %d items, got %d", tt.wantLen, len(resp.Videos))
			}
		})
	}
}

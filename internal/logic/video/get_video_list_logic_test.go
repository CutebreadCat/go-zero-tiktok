package video

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestGetVideoListLogic_GetVideoList(t *testing.T) {
	tests := []struct {
		name    string
		req     *types.GetVideoListRequest
		video   *mock.VideoRepo
		popular *mock.PopularRepo
		wantErr bool
		wantLen int
	}{
		{
			name:    "page size too large",
			req:     &types.GetVideoListRequest{UserID: "u1", PageSize: 200},
			wantErr: true,
		},
		{
			name: "get videos failed",
			req:  &types.GetVideoListRequest{UserID: "u1"},
			video: &mock.VideoRepo{
				GetVideosByAuthorIDFn: func(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success with videos",
			req:  &types.GetVideoListRequest{UserID: "u1", PageSize: 10, PageNum: 1},
			video: &mock.VideoRepo{
				GetVideosByAuthorIDFn: func(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
					return []types.VideoBaseinfo{
						{VideoID: "v1", AuthorID: "u1", Title: "video1"},
						{VideoID: "v2", AuthorID: "u1", Title: "video2"},
					}, 2, nil
				},
			},
			popular: &mock.PopularRepo{
				IncreaseVideoVisitCountFn: func(ctx context.Context, videoID string, delta int64) error {
					return nil
				},
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "success empty",
			req:  &types.GetVideoListRequest{UserID: "u1"},
			video: &mock.VideoRepo{
				GetVideosByAuthorIDFn: func(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
					return nil, 0, nil
				},
			},
			popular: &mock.PopularRepo{
				IncreaseVideoVisitCountFn: func(ctx context.Context, videoID string, delta int64) error {
					return nil
				},
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, tt.video, tt.popular, nil, nil, nil, nil)
			logic := NewGetVideoListLogic(context.Background(), svcCtx)
			resp, err := logic.GetVideoList(tt.req)
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
				t.Errorf("expected %d videos, got %d", tt.wantLen, len(resp.Videos))
			}
		})
	}
}

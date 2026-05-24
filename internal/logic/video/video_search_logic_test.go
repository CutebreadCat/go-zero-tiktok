package video

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestVideoSearchLogic_VideoSearch(t *testing.T) {
	tests := []struct {
		name    string
		req     *types.VideoSearchRequest
		video   *mock.VideoRepo
		popular *mock.PopularRepo
		wantErr bool
		wantLen int
	}{
		{
			name:    "page size too large",
			req:     &types.VideoSearchRequest{PageSize: 200},
			wantErr: true,
		},
		{
			name: "search failed",
			req:  &types.VideoSearchRequest{Keyword: "test"},
			video: &mock.VideoRepo{
				SearchVideosByKeywordFn: func(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success with results",
			req:  &types.VideoSearchRequest{Keyword: "test", PageSize: 10, PageNum: 1},
			video: &mock.VideoRepo{
				SearchVideosByKeywordFn: func(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
					return []types.VideoBaseinfo{
						{VideoID: "v1", Title: "test video"},
					}, 1, nil
				},
			},
			popular: &mock.PopularRepo{
				IncreaseVideoVisitCountFn: func(ctx context.Context, videoID string, delta int64) error {
					return nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "success empty results",
			req:  &types.VideoSearchRequest{Keyword: "notfound"},
			video: &mock.VideoRepo{
				SearchVideosByKeywordFn: func(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
					return nil, 0, nil
				},
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, tt.video, tt.popular, nil, nil, nil, nil)
			logic := NewVideoSearchLogic(context.Background(), svcCtx)
			resp, err := logic.VideoSearch(context.Background(), tt.req)
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

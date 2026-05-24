package interaction

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestGetCommentListLogic_GetCommentList(t *testing.T) {
	tests := []struct {
		name    string
		req     *types.GetCommentListRequest
		comment *mock.CommentRepo
		wantErr bool
		wantLen int
	}{
		{
			name:    "empty video id",
			req:     &types.GetCommentListRequest{VideoID: ""},
			wantErr: true,
		},
		{
			name: "page size too large",
			req:  &types.GetCommentListRequest{VideoID: "v1", PageSize: 200},
			wantErr: true,
		},
		{
			name: "get comments failed",
			req:  &types.GetCommentListRequest{VideoID: "v1"},
			comment: &mock.CommentRepo{
				GetCommentsByVideoIDFn: func(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success with comments",
			req:  &types.GetCommentListRequest{VideoID: "v1", PageSize: 10, PageNumber: 1},
			comment: &mock.CommentRepo{
				GetCommentsByVideoIDFn: func(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
					return []types.CommentBaseinfo{
						{CommentID: "c1", UserID: "u1", Content: "test1"},
						{CommentID: "c2", UserID: "u2", Content: "test2"},
					}, 2, nil
				},
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "success empty",
			req:  &types.GetCommentListRequest{VideoID: "v1"},
			comment: &mock.CommentRepo{
				GetCommentsByVideoIDFn: func(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
					return nil, 0, nil
				},
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, tt.comment, nil, nil, nil)
			logic := NewGetCommentListLogic(context.Background(), svcCtx)
			resp, err := logic.GetCommentList(tt.req)
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
			if len(resp.CommentList) != tt.wantLen {
				t.Errorf("expected %d comments, got %d", tt.wantLen, len(resp.CommentList))
			}
		})
	}
}

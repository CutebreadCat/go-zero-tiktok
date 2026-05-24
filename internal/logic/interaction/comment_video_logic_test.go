package interaction

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestCommentVideoLogic_CommentVideo(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		req       *types.CommentVideoRequest
		comment   *mock.CommentRepo
		popular   *mock.PopularRepo
		wantErr   bool
	}{
		{
			name:    "empty video id",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.CommentVideoRequest{VideoID: "", CommentText: "test"},
			wantErr: true,
		},
		{
			name:    "empty comment text",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.CommentVideoRequest{VideoID: "v1", CommentText: ""},
			wantErr: true,
		},
		{
			name: "create comment failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.CommentVideoRequest{VideoID: "v1", CommentText: "test"},
			comment: &mock.CommentRepo{
				CreateCommentFromParamsFn: func(ctx context.Context, commentID, userID, videoID, content, parentCommentID string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.CommentVideoRequest{VideoID: "v1", CommentText: "test comment"},
			comment: &mock.CommentRepo{
				CreateCommentFromParamsFn: func(ctx context.Context, commentID, userID, videoID, content, parentCommentID string) error {
					return nil
				},
			},
			popular: &mock.PopularRepo{
				IncreaseVideoVisitCountFn: func(ctx context.Context, videoID string, delta int64) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, tt.popular, tt.comment, nil, nil, nil)
			logic := NewCommentVideoLogic(tt.ctx, svcCtx)
			resp, err := logic.CommentVideo(tt.req)
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
			if resp.CommentID == "" {
				t.Errorf("expected comment ID, got empty")
			}
		})
	}
}

package interaction

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestLikeCommentLogic_LikeComment(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		req     *types.LikeCommentRequest
		comment *mock.CommentRepo
		wantErr bool
	}{
		{
			name:    "invalid like type",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.LikeCommentRequest{CommentID: "c1", Liketype: 2},
			wantErr: true,
		},
		{
			name: "like failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.LikeCommentRequest{CommentID: "c1", Liketype: 1},
			comment: &mock.CommentRepo{
				LikeCommentFn: func(ctx context.Context, commentID string, userID string, likeType int32) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success like",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.LikeCommentRequest{CommentID: "c1", Liketype: 1},
			comment: &mock.CommentRepo{
				LikeCommentFn: func(ctx context.Context, commentID string, userID string, likeType int32) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "success unlike",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.LikeCommentRequest{CommentID: "c1", Liketype: 0},
			comment: &mock.CommentRepo{
				LikeCommentFn: func(ctx context.Context, commentID string, userID string, likeType int32) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, tt.comment, nil, nil, nil)
			logic := NewLikeCommentLogic(tt.ctx, svcCtx)
			_, err := logic.LikeComment(tt.req)
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

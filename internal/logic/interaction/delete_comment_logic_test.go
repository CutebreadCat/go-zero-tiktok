package interaction

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestDeleteCommentLogic_DeleteComment(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		req     *types.DeleteCommentRequest
		comment *mock.CommentRepo
		wantErr bool
	}{
		{
			name:    "empty comment id",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.DeleteCommentRequest{CommentID: ""},
			wantErr: true,
		},
		{
			name: "delete failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.DeleteCommentRequest{CommentID: "c1"},
			comment: &mock.CommentRepo{
				DeleteCommentByIDFn: func(ctx context.Context, commentID string, userID string) error {
					return errors.New("not found")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.DeleteCommentRequest{CommentID: "c1"},
			comment: &mock.CommentRepo{
				DeleteCommentByIDFn: func(ctx context.Context, commentID string, userID string) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, tt.comment, nil, nil, nil)
			logic := NewDeleteCommentLogic(tt.ctx, svcCtx)
			_, err := logic.DeleteComment(tt.req)
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

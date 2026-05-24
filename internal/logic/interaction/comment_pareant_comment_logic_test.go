package interaction

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestCommentPareantCommentLogic_CommentPareantComment(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		req     *types.CommentPareantCommentRequest
		comment *mock.CommentRepo
		wantErr bool
	}{
		{
			name: "comment parent failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.CommentPareantCommentRequest{ParentCommentID: "c1", CommentText: "reply"},
			comment: &mock.CommentRepo{
				CommentParentComentFn: func(ctx context.Context, userID string, commentText string, parentCommentID string) (string, error) {
					return "", errors.New("parent not found")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.CommentPareantCommentRequest{ParentCommentID: "c1", CommentText: "reply"},
			comment: &mock.CommentRepo{
				CommentParentComentFn: func(ctx context.Context, userID string, commentText string, parentCommentID string) (string, error) {
					return "c2", nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, tt.comment, nil, nil, nil)
			logic := NewCommentPareantCommentLogic(tt.ctx, svcCtx)
			resp, err := logic.CommentPareantComment(tt.req)
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

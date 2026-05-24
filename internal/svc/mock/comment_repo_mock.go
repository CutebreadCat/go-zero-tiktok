package mock

import (
	"context"

	"go_zero-tiktok/internal/types"
)

type CommentRepo struct {
	CreateCommentFromParamsFn func(ctx context.Context, commentID, userID, videoID, content, parentCommentID string) error
	DeleteCommentByIDFn       func(ctx context.Context, commentID string, userID string) error
	GetCommentsByVideoIDFn    func(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error)
	LikeCommentFn             func(ctx context.Context, commentID string, userID string, likeType int32) error
	CommentParentComentFn     func(ctx context.Context, userID string, commentText string, parentCommentID string) (string, error)
}

func (m *CommentRepo) CreateCommentFromParams(ctx context.Context, commentID, userID, videoID, content, parentCommentID string) error {
	if m.CreateCommentFromParamsFn != nil {
		return m.CreateCommentFromParamsFn(ctx, commentID, userID, videoID, content, parentCommentID)
	}
	return nil
}

func (m *CommentRepo) DeleteCommentByID(ctx context.Context, commentID string, userID string) error {
	if m.DeleteCommentByIDFn != nil {
		return m.DeleteCommentByIDFn(ctx, commentID, userID)
	}
	return nil
}

func (m *CommentRepo) GetCommentsByVideoID(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	if m.GetCommentsByVideoIDFn != nil {
		return m.GetCommentsByVideoIDFn(ctx, videoID, pageNumber, pageSize)
	}
	return nil, 0, nil
}

func (m *CommentRepo) LikeComment(ctx context.Context, commentID string, userID string, likeType int32) error {
	if m.LikeCommentFn != nil {
		return m.LikeCommentFn(ctx, commentID, userID, likeType)
	}
	return nil
}

func (m *CommentRepo) CommentParentComent(ctx context.Context, userID string, commentText string, parentCommentID string) (string, error) {
	if m.CommentParentComentFn != nil {
		return m.CommentParentComentFn(ctx, userID, commentText, parentCommentID)
	}
	return "", nil
}

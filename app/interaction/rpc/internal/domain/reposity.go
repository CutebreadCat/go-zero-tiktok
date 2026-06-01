package domain

import (
	"context"
	"go_zero-tiktok/internal/types"
)

type ICommentRepo interface {
	CreateCommentFromParams(ctx context.Context, commentID, userID, videoID, content, parentCommentID string) error
	DeleteCommentByID(ctx context.Context, commentID string, userID string) error
	GetCommentsByVideoID(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error)
	LikeComment(ctx context.Context, commentID string, userID string, likeType int32) error
	CommentParentComent(ctx context.Context, userID string, commentText string, parentCommentID string) (string, error)
}

type IVideoVisitRecorder interface {
	IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error
}

package domain

import (
	"context"
	"go_zero-tiktok/pkg/contract"
)

type ICommentRepo interface {
	CreateCommentFromParams(ctx context.Context, commentID, userID, videoID int64, content string, parentCommentID int64) error
	DeleteCommentByID(ctx context.Context, commentID int64, userID int64) error
	GetCommentsByVideoID(ctx context.Context, videoID int64, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error)
	LikeComment(ctx context.Context, commentID int64, userID int64, likeType int32) error
	CommentParentComent(ctx context.Context, userID int64, commentText string, parentCommentID int64) (int64, error)
}

type IVideoVisitRecorder interface {
	IncreaseVideoVisitCount(ctx context.Context, videoID int64, delta int64) error
}

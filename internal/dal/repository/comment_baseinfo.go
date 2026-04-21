package repository

import (
	"context"

	commenttable "go_zero-tiktok/internal/dal/tables/comment_baseinfo"
	"go_zero-tiktok/internal/types"

	"gorm.io/gorm"
)

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) CreateComment(ctx context.Context, comment *types.CommentBaseinfo) error {
	return commenttable.CreateComment(ctx, r.db, comment)
}

func (r *CommentRepo) DeleteCommentByID(ctx context.Context, commentID string, userID string) error {
	return commenttable.DeleteCommentByID(ctx, r.db, commentID, userID)
}

func (r *CommentRepo) GetCommentsByVideoID(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	return commenttable.GetCommentsByVideoID(ctx, r.db, videoID, pageNumber, pageSize)
}

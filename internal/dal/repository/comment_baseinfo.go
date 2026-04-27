package repository

import (
	"context"
	"log"

	commenttable "go_zero-tiktok/internal/dal/tables/comment_baseinfo"
	"go_zero-tiktok/internal/svc/xerr"
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

func (r *CommentRepo) LikeComment(ctx context.Context, commentID string, userID string, likeType int32) error {
	switch likeType {
	case 1:
		return commenttable.LikeComment(ctx, r.db, commentID, userID)
	case 0:
		return commenttable.UnlikeComment(ctx, r.db, commentID, userID)
	default:
		return nil
	}

}

func (r *CommentRepo) CommentParentComent(ctx context.Context, userID string, commentText string, parentCommentID string) (string, error) {
	if parentCommentID == "" {
		log.Printf("parentCommentID is empty")
		return "", xerr.New(400, "父评论ID不能为空")
	}
	// 验证父评论是否存在
	var parentComment types.CommentBaseinfo
	if err := r.db.WithContext(ctx).Where("comment_id = ?", parentCommentID).First(&parentComment).Error; err != nil {
		log.Printf("parent comment not found: %v", err)
		return "", xerr.New(400, "父评论不存在")
	}
	var commentId string
	var err error
	if commentId, err = commenttable.CommentPareantComment(ctx, r.db, parentCommentID, commentText, userID, parentComment.VideoID); err != nil {
		log.Printf("comment parent comment failed: %v", err)
		return "", err
	}
	return commentId, nil

}

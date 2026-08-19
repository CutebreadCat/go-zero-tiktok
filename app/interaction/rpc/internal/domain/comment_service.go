package domain

import (
	"context"

	"go_zero-tiktok/pkg/contract"
)

type CommentService struct {
	commentRepo ICommentRepo
}

func NewCommentService(commentRepo ICommentRepo) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
	}
}

// CreateComment 创建评论。
func (s *CommentService) CreateComment(ctx context.Context, commentID, userID, videoID int64, content string, parentCommentID int64) error {
	return s.commentRepo.CreateCommentFromParams(ctx, commentID, userID, videoID, content, parentCommentID)
}

// ReplyParentComment 回复父评论
func (s *CommentService) ReplyParentComment(ctx context.Context, userID int64, commentText string, parentCommentID int64) (int64, error) {
	return s.commentRepo.CommentParentComment(ctx, userID, commentText, parentCommentID)
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(ctx context.Context, commentID, userID int64) error {
	return s.commentRepo.DeleteCommentByID(ctx, commentID, userID)
}

// LikeComment 点赞评论
func (s *CommentService) LikeComment(ctx context.Context, commentID, userID int64, likeType int32) error {
	return s.commentRepo.LikeComment(ctx, commentID, userID, likeType)
}

// GetCommentsByVideoID 获取视频评论列表
func (s *CommentService) GetCommentsByVideoID(ctx context.Context, videoID int64, pageNum, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	return s.commentRepo.GetCommentsByVideoID(ctx, videoID, pageNum, pageSize)
}

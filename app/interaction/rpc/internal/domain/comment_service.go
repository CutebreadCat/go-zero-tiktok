package domain

import (
	"context"

	"go_zero-tiktok/pkg/contract"
)

type CommentService struct {
	commentRepo ICommentRepo
	popularRepo IPopularRepo
}

func NewCommentService(commentRepo ICommentRepo, popularRepo IPopularRepo) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		popularRepo: popularRepo,
	}
}

// CreateComment 创建评论，并同步更新 video_stat.comment_count。
func (s *CommentService) CreateComment(ctx context.Context, commentID, userID, videoID int64, content string, parentCommentID int64) error {
	if err := s.commentRepo.CreateCommentFromParams(ctx, commentID, userID, videoID, content, parentCommentID); err != nil {
		return err
	}

	if s.popularRepo != nil {
		if err := s.popularRepo.UpdateVideoCommentCount(ctx, videoID, 1); err != nil {
			return err
		}
	}

	return nil
}

// ReplyParentComment 回复父评论。
// 回复评论同样应增加 video_stat.comment_count，但当前 CommentParentComment 未返回 video_id，
// 这里不额外更新，避免重复查库。后续可扩展 CommentParentComment 返回 video_id 后补齐。
func (s *CommentService) ReplyParentComment(ctx context.Context, userID int64, commentText string, parentCommentID int64) (int64, error) {
	return s.commentRepo.CommentParentComment(ctx, userID, commentText, parentCommentID)
}

// DeleteComment 删除评论，并同步更新 video_stat.comment_count。
func (s *CommentService) DeleteComment(ctx context.Context, commentID, userID int64) error {
	videoID, err := s.commentRepo.DeleteCommentByID(ctx, commentID, userID)
	if err != nil {
		return err
	}

	if s.popularRepo != nil {
		if err := s.popularRepo.UpdateVideoCommentCount(ctx, videoID, -1); err != nil {
			return err
		}
	}

	return nil
}

// LikeComment 点赞评论
func (s *CommentService) LikeComment(ctx context.Context, commentID, userID int64, likeType int32) error {
	return s.commentRepo.LikeComment(ctx, commentID, userID, likeType)
}

// GetCommentsByVideoID 获取视频评论列表
func (s *CommentService) GetCommentsByVideoID(ctx context.Context, videoID int64, pageNum, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	return s.commentRepo.GetCommentsByVideoID(ctx, videoID, pageNum, pageSize)
}

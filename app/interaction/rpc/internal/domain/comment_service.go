package domain

import (
	"context"
	"fmt"
	"time"

	"go_zero-tiktok/pkg/contract"
)

type CommentService struct {
	commentRepo        ICommentRepo
	videoVisitRecorder IVideoVisitRecorder
}

func NewCommentService(commentRepo ICommentRepo, videoVisitRecorder IVideoVisitRecorder) *CommentService {
	return &CommentService{
		commentRepo:        commentRepo,
		videoVisitRecorder: videoVisitRecorder,
	}
}

// CreateComment 创建评论并异步记录访问量
func (s *CommentService) CreateComment(ctx context.Context, commentID, userID, videoID int64, content string, parentCommentID int64) error {
	if err := s.commentRepo.CreateCommentFromParams(ctx, commentID, userID, videoID, content, parentCommentID); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := s.videoVisitRecorder.IncreaseVideoVisitCount(ctx, videoID, 1); err != nil {
			fmt.Printf("increment visit count failed for video %d: %v\n", videoID, err)
		}
	}()

	return nil
}

// ReplyParentComment 回复父评论
func (s *CommentService) ReplyParentComment(ctx context.Context, userID int64, commentText string, parentCommentID int64) (int64, error) {
	return s.commentRepo.CommentParentComent(ctx, userID, commentText, parentCommentID)
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

package repository

import (
	"context"
	"log"

	commenttable "go_zero-tiktok/internal/dal/tables/comment_baseinfo"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"gorm.io/gorm"
)

// CommentToResponse 将数据库模型转换为API响应类型
func (r *CommentRepo) CommentToResponse(comment *commenttable.CommentBaseinfo) types.CommentBaseinfo {
	return types.CommentBaseinfo{
		CommentID:       comment.CommentID,
		UserID:          comment.UserID,
		VideoID:         comment.VideoID,
		Content:         comment.Content,
		ParentCommentID: comment.ParentCommentID,
		CreatedAt:       myutils.TimeToStr(comment.CreatedAt, ""),
		UpdatedAt:       myutils.TimeToStr(comment.UpdatedAt, ""),
	}
}

// CommentsToResponse 将数据库模型切片转换为API响应类型切片
func (r *CommentRepo) CommentsToResponse(comments []commenttable.CommentBaseinfo) []types.CommentBaseinfo {
	result := make([]types.CommentBaseinfo, 0, len(comments))
	for _, c := range comments {
		result = append(result, r.CommentToResponse(&c))
	}
	return result
}

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) CreateComment(ctx context.Context, comment *commenttable.CommentBaseinfo) error {
	return commenttable.CreateComment(ctx, r.db, comment)
}

// CreateCommentFromParams 通过参数创建评论，logic层不需要知道数据库模型
func (r *CommentRepo) CreateCommentFromParams(ctx context.Context, commentID, userID, videoID, content, parentCommentID string) error {
	comment := &commenttable.CommentBaseinfo{
		CommentID:       commentID,
		UserID:          userID,
		VideoID:         videoID,
		Content:         content,
		ParentCommentID: parentCommentID,
	}
	return commenttable.CreateComment(ctx, r.db, comment)
}

func (r *CommentRepo) DeleteCommentByID(ctx context.Context, commentID string, userID string) error {
	return commenttable.DeleteCommentByID(ctx, r.db, commentID, userID)
}

func (r *CommentRepo) GetCommentsByVideoID(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]commenttable.CommentBaseinfo, int64, error) {
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
	var parentComment commenttable.CommentBaseinfo
	if err := r.db.Model(&commenttable.CommentBaseinfo{}).Where("comment_id = ?", parentCommentID).First(&parentComment).Error; err != nil {
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

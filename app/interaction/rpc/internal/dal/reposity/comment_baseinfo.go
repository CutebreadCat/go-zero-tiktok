package reposity

import (
	"context"
	"errors"

	commenttable "go_zero-tiktok/app/interaction/rpc/internal/dal/tables/comment_baseinfo"
	"go_zero-tiktok/pkg/contract"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

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

func (r *CommentRepo) CommentsToResponse(comments []commenttable.CommentBaseinfo) []types.CommentBaseinfo {
	result := make([]types.CommentBaseinfo, 0, len(comments))
	for _, c := range comments {
		result = append(result, r.CommentToResponse(&c))
	}
	return result
}

func (r *CommentRepo) CreateComment(ctx context.Context, comment *commenttable.CommentBaseinfo) error {
	if err := commenttable.CreateComment(ctx, r.db, comment); err != nil {
		return pkgerrors.WithMessage(err, "CommentRepo.CreateComment")
	}
	return nil
}

func (r *CommentRepo) CreateCommentFromParams(ctx context.Context, commentID, userID, videoID int64, content string, parentCommentID int64) error {
	comment := &commenttable.CommentBaseinfo{
		CommentID:       commentID,
		UserID:          userID,
		VideoID:         videoID,
		Content:         content,
		ParentCommentID: parentCommentID,
	}
	if err := commenttable.CreateComment(ctx, r.db, comment); err != nil {
		return pkgerrors.WithMessage(err, "CommentRepo.CreateCommentFromParams")
	}
	return nil
}

func (r *CommentRepo) DeleteCommentByID(ctx context.Context, commentID int64, userID int64) error {
	if err := commenttable.DeleteCommentByID(ctx, r.db, commentID, userID); err != nil {
		return pkgerrors.WithMessage(err, "CommentRepo.DeleteCommentByID")
	}
	return nil
}

func (r *CommentRepo) GetCommentsByVideoID(ctx context.Context, videoID int64, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	comments, total, err := commenttable.GetCommentsByVideoID(ctx, r.db, videoID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "CommentRepo.GetCommentsByVideoID")
	}
	return r.CommentsToResponse(comments), total, nil
}

func (r *CommentRepo) LikeComment(ctx context.Context, commentID int64, userID int64, likeType int32) error {
	switch likeType {
	case 1:
		if err := commenttable.LikeComment(ctx, r.db, commentID, userID); err != nil {
			return pkgerrors.WithMessage(err, "CommentRepo.LikeComment")
		}
		return nil
	case 0:
		if err := commenttable.UnlikeComment(ctx, r.db, commentID, userID); err != nil {
			return pkgerrors.WithMessage(err, "CommentRepo.LikeComment:unlike")
		}
		return nil
	default:
		return nil
	}
}

func (r *CommentRepo) CommentParentComent(ctx context.Context, userID int64, commentText string, parentCommentID int64) (int64, error) {
	if parentCommentID == 0 {
		return 0, xerr.NewInvalidParam("父评论ID不能为空")
	}

	var parentComment commenttable.CommentBaseinfo
	if err := r.db.Model(&commenttable.CommentBaseinfo{}).Where("comment_id = ?", parentCommentID).First(&parentComment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, xerr.NewInvalidParam("父评论不存在")
		}
		return 0, xerr.Wrap(err, "query parent comment failed")
	}

	commentId, err := commenttable.CommentPareantComment(ctx, r.db, parentCommentID, commentText, userID, parentComment.VideoID)
	if err != nil {
		return 0, pkgerrors.WithMessage(err, "CommentRepo.CommentParentComent")
	}
	return commentId, nil
}

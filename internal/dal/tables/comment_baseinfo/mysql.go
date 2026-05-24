package comment_baseinfo

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/shared/xerr"
	myutils "go_zero-tiktok/internal/utils"

	"gorm.io/gorm"
)

func CreateComment(ctx context.Context, db *gorm.DB, comment *CommentBaseinfo) error {
	if comment == nil {
		return xerr.NewInvalidParam("评论不存在")
	}

	if err := db.WithContext(ctx).Create(comment).Error; err != nil {
		return xerr.Wrap(err, "create comment failed")
	}

	return nil
}

func DeleteCommentByID(ctx context.Context, db *gorm.DB, commentID string, userID string) error {
	var comment CommentBaseinfo
	if err := db.WithContext(ctx).Where("comment_id = ?", commentID).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xerr.NewInvalidParam("评论不存在")
		}
		return xerr.Wrap(err, "delete comment query failed")
	}
	if comment.UserID != userID {
		return xerr.NewInvalidParam("删除评论失败，用户ID不匹配")
	}

	result := db.WithContext(ctx).Where("comment_id = ?", commentID).Delete(&comment)
	if result.Error != nil {
		return xerr.Wrap(result.Error, "delete comment failed")
	}

	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("删除评论失败")
	}

	return nil
}

func GetCommentsByVideoID(ctx context.Context, db *gorm.DB, videoID string, pageNumber, pageSize int32) ([]CommentBaseinfo, int64, error) {
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&CommentBaseinfo{}).Where("video_id = ?", videoID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get comments by video id count failed")
	}

	var comments []CommentBaseinfo
	offset := (pageNumber - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&comments).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get comments by video id failed")
	}

	return comments, total, nil
}

func LikeComment(ctx context.Context, db *gorm.DB, commentID string, userID string) error {
	like := CommentLiker{
		UserID:    userID,
		CommentID: commentID,
	}

	if err := db.WithContext(ctx).Create(&like).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return xerr.NewInvalidParam("已经点赞过该评论了")
		}
		return xerr.Wrap(err, "like comment failed")
	}

	return nil
}

func UnlikeComment(ctx context.Context, db *gorm.DB, commentID string, userID string) error {
	result := db.WithContext(ctx).Where("user_id = ? AND comment_id = ?", userID, commentID).Delete(&CommentLiker{})
	if result.Error != nil {
		return xerr.Wrap(result.Error, "unlike comment failed")
	}

	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("取消点赞评论失败,没有找到点赞记录")
	}
	return nil
}

func CommentPareantComment(ctx context.Context, db *gorm.DB, parentCommentID string, commentText string, userID string, videoID string) (string, error) {
	comment := &CommentBaseinfo{
		CommentID:       myutils.GenerateCommentID(),
		UserID:          userID,
		VideoID:         videoID,
		Content:         commentText,
		ParentCommentID: parentCommentID,
	}

	if err := db.WithContext(ctx).Create(comment).Error; err != nil {
		return "", xerr.Wrap(err, "create comment failed")
	}

	return comment.CommentID, nil
}

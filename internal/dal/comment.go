package dal

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func CreateComment(ctx context.Context, comment *types.CommentBaseinfo) error {
	logger := logx.WithContext(ctx)

	if comment == nil {
		err := errors.New("comment is nil")
		logger.Errorf("create comment failed: %v", err)
		return err
	}

	if err := Db.WithContext(ctx).Create(comment).Error; err != nil {
		logger.Errorf("create comment failed: %v", err)
		return err
	}

	return nil
}

func DeleteCommentByID(ctx context.Context, commentID string) error {
	logger := logx.WithContext(ctx)

	result := Db.WithContext(ctx).Where("comment_id = ?", commentID).Delete(&types.CommentBaseinfo{})
	if result.Error != nil {
		logger.Errorf("delete comment failed: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("delete comment failed: %v", err)
		return err
	}

	return nil
}

func GetCommentsByVideoID(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := Db.WithContext(ctx).Model(&types.CommentBaseinfo{}).Where("video_id = ?", videoID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get comments by video id count failed: %v", err)
		return nil, 0, err
	}

	var comments []types.CommentBaseinfo
	offset := (pageNumber - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&comments).Error; err != nil {
		logger.Errorf("get comments by video id failed: %v", err)
		return nil, 0, err
	}

	return comments, total, nil
}

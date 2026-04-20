package comment_baseinfo

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func CreateComment(ctx context.Context, db *gorm.DB, comment *types.CommentBaseinfo) error {
	logger := logx.WithContext(ctx)

	if comment == nil {
		err := errors.New("comment is nil")
		logger.Errorf("create comment failed: %v", err)
		return xerr.New(400, "评论不存在")
	}

	if err := db.WithContext(ctx).Create(comment).Error; err != nil {
		logger.Errorf("create comment failed: %v", err)
		return errors.New("创建评论失败")
	}

	return nil
}

func DeleteCommentByID(ctx context.Context, db *gorm.DB, commentID string) error {
	logger := logx.WithContext(ctx)

	result := db.WithContext(ctx).Where("comment_id = ?", commentID).Delete(&types.CommentBaseinfo{})
	if result.Error != nil {
		logger.Errorf("delete comment failed: %v", result.Error)
		return errors.New("删除评论失败")
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("delete comment failed: %v", err)
		return errors.New("删除评论失败")
	}

	return nil
}

func GetCommentsByVideoID(ctx context.Context, db *gorm.DB, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.CommentBaseinfo{}).Where("video_id = ?", videoID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get comments by video id count failed: %v", err)
		return nil, 0, errors.New("获取评论总数失败")
	}

	var comments []types.CommentBaseinfo
	offset := (pageNumber - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&comments).Error; err != nil {
		logger.Errorf("get comments by video id failed: %v", err)
		return nil, 0, errors.New("获取评论失败")
	}

	return comments, total, nil
}

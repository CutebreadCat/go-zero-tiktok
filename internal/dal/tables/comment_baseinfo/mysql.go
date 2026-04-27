package comment_baseinfo

import (
	"context"
	"errors"
	"time"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"fmt"

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
		return xerr.New(400, "创建评论失败")
	}

	return nil
}

func DeleteCommentByID(ctx context.Context, db *gorm.DB, commentID string, userID string) error {
	logger := logx.WithContext(ctx)
	var comment types.CommentBaseinfo
	if err := db.WithContext(ctx).Where("comment_id = ?", commentID).First(&comment).Error; err != nil {
		logger.Errorf("delete comment failed: %v", err)
		return xerr.New(400, "评论不存在")
	}
	if comment.UserID != userID {
		logger.Errorf("delete comment failed: %v", fmt.Sprintf("user_id not match: %s, expect %s", comment.UserID, userID))
		return xerr.New(400, "删除评论失败，用户ID不匹配")
	}

	result := db.WithContext(ctx).Where("comment_id = ?", commentID).Delete(&comment)
	if result.Error != nil {
		logger.Errorf("delete comment failed: %v", result.Error)
		return xerr.New(400, "删除评论失败")
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("delete comment failed: %v", err)
		return xerr.New(400, "删除评论失败")

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
		return nil, 0, xerr.New(400, "获取评论总数失败")
	}

	var comments []types.CommentBaseinfo
	offset := (pageNumber - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&comments).Error; err != nil {
		logger.Errorf("get comments by video id failed: %v", err)
		return nil, 0, xerr.New(400, "获取评论失败")
	}

	return comments, total, nil
}
func LikeComment(ctx context.Context, db *gorm.DB, commentID string, userID string) error {
	logger := logx.WithContext(ctx)

	like := types.CommentLiker{
		UserID:    userID,
		CommentID: commentID,
	}

	if err := db.WithContext(ctx).Create(&like).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Errorf("like comment failed: user %s already liked comment %s", userID, commentID)
			return xerr.New(400, "已经点赞过该评论了")
		}
		logger.Errorf("like comment failed: %v", err)
		return xerr.New(400, "点赞评论失败,服务器内部出现错误")
	}

	return nil

}

func UnlikeComment(ctx context.Context, db *gorm.DB, commentID string, userID string) error {
	logger := logx.WithContext(ctx)
	result := db.WithContext(ctx).Where("user_id = ? AND comment_id = ?", userID, commentID).Delete(&types.CommentLiker{})
	if result.Error != nil {
		logger.Errorf("unlike comment failed: %v", result.Error)
		return xerr.New(400, "取消点赞评论失败,服务器内部出现错误")
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("unlike comment failed: %v", err)
		return xerr.New(400, "取消点赞评论失败,没有找到点赞记录")
	}
	return nil
}
func CommentPareantComment(ctx context.Context, db *gorm.DB, parentCommentID string, commentText string, userID string, videoID string) (string, error) {
	logger := logx.WithContext(ctx)

	comment := &types.CommentBaseinfo{
		CommentID:       myutils.GenerateCommentID(),
		UserID:          userID,
		VideoID:         videoID,
		Content:         commentText,
		ParentCommentID: parentCommentID,
		CreatedAt:       myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05"),
		UpdatedAt:       myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05"),
		DeletedAt:       "",
	}

	if err := db.WithContext(ctx).Create(comment).Error; err != nil {
		logger.Errorf("create comment failed: %v", err)
		return "", xerr.New(400, "创建评论失败")
	}

	return "", nil
}

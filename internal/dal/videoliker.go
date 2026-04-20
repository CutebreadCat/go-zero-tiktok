package dal

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	userLikedVideosSetPrefix = "user:liked:videos:"
)

func userLikedVideosSetKey(userID string) string {
	return userLikedVideosSetPrefix + userID
}

func LikeVideo(ctx context.Context, userID, videoID string) error {
	logger := logx.WithContext(ctx)

	if userID == "" || videoID == "" {
		err := errors.New("userID or videoID is empty")
		logger.Errorf("like video failed: %v", err)
		return xerr.New(400, "用户ID或视频ID为空")
	}

	like := &types.VideoLiker{
		UserID:  userID,
		VideoID: videoID,
	}

	if err := Db.WithContext(ctx).Create(like).Error; err != nil {
		logger.Errorf("like video failed: %v", err)
		return xerr.New(1002, "点赞视频失败")
	}

	// 数据库为主，Redis 作为缓存加速，缓存写失败不阻断主流程。
	if err := addVideoLikeToRedis(ctx, userID, videoID); err != nil {
		logger.Errorf("sync like to redis failed: %v", err)
	}

	return nil
}

func CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	logger := logx.WithContext(ctx)

	result := Db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&types.VideoLiker{})
	if result.Error != nil {
		logger.Errorf("cancel like video failed: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		err := gorm.ErrRecordNotFound
		logger.Errorf("cancel like video failed: %v", err)
		return xerr.New(400, "点赞关系不存在")
	}

	// 数据库删除成功后再清理缓存，缓存失败不影响主流程。
	if err := removeVideoLikeFromRedis(ctx, userID, videoID); err != nil {
		logger.Errorf("sync unlike to redis failed: %v", err)
	}

	return nil
}

func addVideoLikeToRedis(ctx context.Context, userID, videoID string) error {
	if Rdb == nil {
		logx.WithContext(ctx).Error("Redis client is not initialized")
		return xerr.ErrRedisError
	}

	if _, err := Rdb.SaddCtx(ctx, userLikedVideosSetKey(userID), videoID); err != nil {
		logx.WithContext(ctx).Errorf("add video like to redis failed: %v", err)
		return err
	}

	return nil
}

func removeVideoLikeFromRedis(ctx context.Context, userID, videoID string) error {
	if Rdb == nil {
		return xerr.ErrRedisError
	}

	if _, err := Rdb.SremCtx(ctx, userLikedVideosSetKey(userID), videoID); err != nil {
		return err
	}

	return nil
}

func GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := Db.WithContext(ctx).Model(&types.VideoLiker{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get liked video ids count failed: %v", err)
		return nil, 0, xerr.New(1002, "获取点赞视频ID总数失败")
	}

	var likerRows []types.VideoLiker
	offset := (pageNumber - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&likerRows).Error; err != nil {
		logger.Errorf("get liked video ids failed: %v", err)
		return nil, 0, xerr.New(1002, "获取点赞视频ID失败")
	}

	videoIDs := make([]string, 0, len(likerRows))
	for _, row := range likerRows {
		videoIDs = append(videoIDs, row.VideoID)
	}

	return videoIDs, total, nil
}

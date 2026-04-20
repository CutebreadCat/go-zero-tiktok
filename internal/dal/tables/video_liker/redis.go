package video_liker

import (
	"context"

	"go_zero-tiktok/internal/svc/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const userLikedVideosSetPrefix = "user:liked:videos:"

func userLikedVideosSetKey(userID string) string {
	return userLikedVideosSetPrefix + userID
}

func AddVideoLike(ctx context.Context, rdb *redis.Redis, userID, videoID string) error {
	if rdb == nil {
		logx.WithContext(ctx).Error("Redis client is not initialized")
		return xerr.ErrRedisError
	}

	if _, err := rdb.SaddCtx(ctx, userLikedVideosSetKey(userID), videoID); err != nil {
		logx.WithContext(ctx).Errorf("add video like to redis failed: %v", err)
		return err
	}

	return nil
}

func RemoveVideoLike(ctx context.Context, rdb *redis.Redis, userID, videoID string) error {
	if rdb == nil {
		return xerr.ErrRedisError
	}

	if _, err := rdb.SremCtx(ctx, userLikedVideosSetKey(userID), videoID); err != nil {
		return err
	}

	return nil
}

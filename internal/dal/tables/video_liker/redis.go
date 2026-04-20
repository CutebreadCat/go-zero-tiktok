package video_liker

import (
	"context"
	"sort"

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

func GetLikedVideoIDs(ctx context.Context, rdb *redis.Redis, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	if rdb == nil {
		return nil, 0, xerr.ErrRedisError
	}

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	all, err := rdb.SmembersCtx(ctx, userLikedVideosSetKey(userID))
	if err != nil {
		return nil, 0, err
	}
	if len(all) == 0 {
		return nil, 0, nil
	}

	// Redis Set is unordered; sort to keep pagination stable.
	sort.Strings(all)
	total := int64(len(all))

	offset := int((pageNumber - 1) * pageSize)
	if offset >= len(all) {
		return []string{}, total, nil
	}
	end := offset + int(pageSize)
	if end > len(all) {
		end = len(all)
	}

	return all[offset:end], total, nil
}

func ResetLikedVideoIDs(ctx context.Context, rdb *redis.Redis, userID string, videoIDs []string) error {
	if rdb == nil {
		return xerr.ErrRedisError
	}

	key := userLikedVideosSetKey(userID)
	if _, err := rdb.DelCtx(ctx, key); err != nil {
		return err
	}
	if len(videoIDs) == 0 {
		return nil
	}

	values := make([]any, 0, len(videoIDs))
	for _, id := range videoIDs {
		values = append(values, id)
	}
	if _, err := rdb.SaddCtx(ctx, key, values...); err != nil {
		return err
	}

	return nil
}

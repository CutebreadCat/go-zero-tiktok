package video_liker

import (
	"context"
	"sort"

	"go_zero-tiktok/internal/dal/page"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func userLikedVideosSetKey(userID string) string {
	return userLikedVideosSetPrefix + userID
}

func AddVideoLike(ctx context.Context, rdb *redis.Redis, userID, videoID string) error {
	if rdb == nil {
		return xerr.Wrap(xerr.ErrRedisError, "redis client is not initialized")
	}

	if _, err := rdb.SaddCtx(ctx, userLikedVideosSetKey(userID), videoID); err != nil {
		return xerr.Wrap(err, "add video like to redis failed")
	}

	return nil
}

func RemoveVideoLike(ctx context.Context, rdb *redis.Redis, userID, videoID string) error {
	if rdb == nil {
		return xerr.Wrap(xerr.ErrRedisError, "redis client is not initialized")
	}

	if _, err := rdb.SremCtx(ctx, userLikedVideosSetKey(userID), videoID); err != nil {
		return xerr.Wrap(err, "remove video like from redis failed")
	}

	return nil
}

func GetLikedVideoIDs(ctx context.Context, rdb *redis.Redis, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	if rdb == nil {
		return nil, 0, xerr.Wrap(xerr.ErrRedisError, "redis client is not initialized")
	}

	if pageNumber <= 0 {
		pageNumber = int32(page.DefaultPageNum)
	}
	if pageSize <= 0 {
		pageSize = int32(page.DefaultPageSize)
	}

	all, err := rdb.SmembersCtx(ctx, userLikedVideosSetKey(userID))
	if err != nil {
		return nil, 0, xerr.Wrap(err, "get liked video ids from redis failed")
	}
	if len(all) == 0 {
		return nil, 0, nil
	}

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
		return xerr.Wrap(xerr.ErrRedisError, "redis client is not initialized")
	}

	key := userLikedVideosSetKey(userID)
	if _, err := rdb.DelCtx(ctx, key); err != nil {
		return xerr.Wrap(err, "reset liked video ids in redis failed")
	}
	if len(videoIDs) == 0 {
		return nil
	}

	values := make([]any, 0, len(videoIDs))
	for _, id := range videoIDs {
		values = append(values, id)
	}
	if _, err := rdb.SaddCtx(ctx, key, values...); err != nil {
		return xerr.Wrap(err, "reset liked video ids in redis failed")
	}

	return nil
}

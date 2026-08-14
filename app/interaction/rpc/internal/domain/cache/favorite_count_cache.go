package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	appLogger "go_zero-tiktok/pkg/logger"
	"go_zero-tiktok/pkg/xerr"
)

// 收藏计数与关系缓存方法。
// 与点赞共用 LikeCountCache 结构体，key 定义见 keys.go。

// FavoriteVideo 记录用户收藏关系，并累加计数。
// 返回 true 表示是新收藏；false 表示已经收藏过（幂等）。
func (c *LikeCountCache) FavoriteVideo(ctx context.Context, userID, videoID int64) (bool, error) {
	usersKey := fmtVideoFavoriteUsersKey(videoID)
	userVideosKey := fmtUserFavoriteVideosKey(userID)
	countField := strconv.FormatInt(videoID, 10)
	member := strconv.FormatInt(userID, 10)
	score := float64(time.Now().UnixMilli())

	script := `
		local added = redis.call('SADD', KEYS[1], ARGV[1])
		if added == 0 then
			return 0
		end
		redis.call('ZADD', KEYS[2], ARGV[2], ARGV[1])
		local count = redis.call('HINCRBY', KEYS[3], ARGV[3], 1)
		if count < 0 then
			redis.call('HSET', KEYS[3], ARGV[3], 0)
			count = 0
		end
		redis.call('SADD', KEYS[4], ARGV[3])
		redis.call('EXPIRE', KEYS[1], ARGV[4])
		redis.call('EXPIRE', KEYS[2], ARGV[4])
		redis.call('EXPIRE', KEYS[3], ARGV[4])
		return 1
	`
	ret, err := c.rdb.EvalCtx(ctx, script,
		[]string{usersKey, userVideosKey, FavoriteCountKey, FavoriteDirtyKey},
		[]any{member, score, countField, likeKeyTTLSeconds},
	)
	if err != nil {
		return false, xerr.Wrap(err, "LikeCountCache.FavoriteVideo")
	}

	v, ok := ret.(int64)
	if !ok {
		v = 0
	}
	return v == 1, nil
}

// CancelFavoriteVideo 移除用户收藏关系，并递减计数。
// 返回 true 表示取消成功；false 表示原本未收藏。
func (c *LikeCountCache) CancelFavoriteVideo(ctx context.Context, userID, videoID int64) (bool, error) {
	usersKey := fmtVideoFavoriteUsersKey(videoID)
	userVideosKey := fmtUserFavoriteVideosKey(userID)
	countField := strconv.FormatInt(videoID, 10)
	member := strconv.FormatInt(userID, 10)

	script := `
		local removed = redis.call('SREM', KEYS[1], ARGV[1])
		if removed == 0 then
			return 0
		end
		redis.call('ZREM', KEYS[2], ARGV[1])
		local count = redis.call('HINCRBY', KEYS[3], ARGV[2], -1)
		if count < 0 then
			redis.call('HSET', KEYS[3], ARGV[2], 0)
			count = 0
		end
		redis.call('SADD', KEYS[4], ARGV[2])
		redis.call('EXPIRE', KEYS[1], ARGV[3])
		redis.call('EXPIRE', KEYS[2], ARGV[3])
		redis.call('EXPIRE', KEYS[3], ARGV[3])
		return 1
	`
	ret, err := c.rdb.EvalCtx(ctx, script,
		[]string{usersKey, userVideosKey, FavoriteCountKey, FavoriteDirtyKey},
		[]any{member, countField, likeKeyTTLSeconds},
	)
	if err != nil {
		return false, xerr.Wrap(err, "LikeCountCache.CancelFavoriteVideo")
	}

	v, ok := ret.(int64)
	if !ok {
		v = 0
	}
	return v == 1, nil
}

// GetFavoritedVideoIDs 分页获取用户收藏的视频 ID 列表（按收藏时间倒序）。
func (c *LikeCountCache) GetFavoritedVideoIDs(ctx context.Context, userID int64, pageNum, pageSize int32) ([]int64, int64, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	userVideosKey := fmtUserFavoriteVideosKey(userID)
	totalInt, err := c.rdb.ZcardCtx(ctx, userVideosKey)
	if err != nil {
		return nil, 0, xerr.Wrap(err, "LikeCountCache.GetFavoritedVideoIDs.Zcard")
	}
	total := int64(totalInt)

	start := int64((pageNum - 1) * pageSize)
	stop := start + int64(pageSize) - 1
	members, err := c.rdb.ZrevrangeCtx(ctx, userVideosKey, start, stop)
	if err != nil {
		return nil, 0, xerr.Wrap(err, "LikeCountCache.GetFavoritedVideoIDs.Zrevrange")
	}

	ids := make([]int64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseInt(strings.TrimSpace(m), 10, 64)
		if err != nil {
			appLogger.Warnf("parse favorited video_id failed: %s", m)
			continue
		}
		ids = append(ids, id)
	}

	return ids, total, nil
}

// GetVideoFavoriteUserIDs 获取收藏某视频的全部用户 ID（供 syncer 同步关系用）。
func (c *LikeCountCache) GetVideoFavoriteUserIDs(ctx context.Context, videoID int64) ([]int64, error) {
	usersKey := fmtVideoFavoriteUsersKey(videoID)
	members, err := c.rdb.SmembersCtx(ctx, usersKey)
	if err != nil {
		return nil, xerr.Wrap(err, "LikeCountCache.GetVideoFavoriteUserIDs")
	}

	ids := make([]int64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseInt(strings.TrimSpace(m), 10, 64)
		if err != nil {
			appLogger.Warnf("parse favorite user_id failed: %s", m)
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetFavoriteCounts 批量获取视频 favorite_count。
// 返回值：命中缓存的计数、未命中的视频 ID 列表。
func (c *LikeCountCache) GetFavoriteCounts(ctx context.Context, videoIDs []int64) (map[int64]int64, []int64, error) {
	if len(videoIDs) == 0 {
		return map[int64]int64{}, nil, nil
	}

	fields := make([]string, 0, len(videoIDs))
	for _, id := range videoIDs {
		fields = append(fields, strconv.FormatInt(id, 10))
	}

	values, err := c.rdb.HmgetCtx(ctx, FavoriteCountKey, fields...)
	if err != nil {
		return nil, nil, xerr.Wrap(err, "LikeCountCache.GetFavoriteCounts.Hmget")
	}

	cached := make(map[int64]int64, len(videoIDs))
	var missed []int64
	for i, id := range videoIDs {
		if i >= len(values) || values[i] == "" {
			missed = append(missed, id)
			continue
		}
		v, err := strconv.ParseInt(values[i], 10, 64)
		if err != nil {
			appLogger.Warnf("parse favorite_count cache failed, video_id=%d value=%s: %v", id, values[i], err)
			missed = append(missed, id)
			continue
		}
		cached[id] = v
	}

	return cached, missed, nil
}

// SetFavoriteCounts 批量写入 favorite_count 缓存。
func (c *LikeCountCache) SetFavoriteCounts(ctx context.Context, counts map[int64]int64) error {
	if len(counts) == 0 {
		return nil
	}

	kvs := make(map[string]string, len(counts))
	for id, count := range counts {
		kvs[strconv.FormatInt(id, 10)] = strconv.FormatInt(count, 10)
	}

	if err := c.rdb.HmsetCtx(ctx, FavoriteCountKey, kvs); err != nil {
		return xerr.Wrap(err, "LikeCountCache.SetFavoriteCounts")
	}
	_ = c.rdb.ExpireCtx(ctx, FavoriteCountKey, likeKeyTTLSeconds)
	return nil
}

// SetFavoriteCount 覆盖指定视频的 favorite_count 当前总值，同时移除脏标记。
func (c *LikeCountCache) SetFavoriteCount(ctx context.Context, videoID int64, count int64) error {
	field := strconv.FormatInt(videoID, 10)

	script := `
		redis.call('HSET', KEYS[1], ARGV[1], tonumber(ARGV[2]))
		redis.call('SREM', KEYS[2], ARGV[1])
		redis.call('EXPIRE', KEYS[1], ARGV[3])
		return 1
	`
	_, err := c.rdb.EvalCtx(ctx, script, []string{FavoriteCountKey, FavoriteDirtyKey}, []any{field, count, likeKeyTTLSeconds})
	if err != nil {
		return xerr.Wrap(err, "LikeCountCache.SetFavoriteCount")
	}
	return nil
}

// PopFavoriteDirtyVideos 从收藏脏集合中随机取出 batch 个待同步视频 ID。
func (c *LikeCountCache) PopFavoriteDirtyVideos(ctx context.Context, batch int) ([]int64, error) {
	if batch <= 0 {
		batch = 100
	}

	members, err := c.rdb.SrandmemberCtx(ctx, FavoriteDirtyKey, batch)
	if err != nil {
		return nil, xerr.Wrap(err, "LikeCountCache.PopFavoriteDirtyVideos.Srandmember")
	}

	if len(members) == 0 {
		return nil, nil
	}

	_, err = c.rdb.SremCtx(ctx, FavoriteDirtyKey, toAnySlice(members)...)
	if err != nil {
		appLogger.Warnf("remove favorite dirty videos failed: %v", err)
	}

	ids := make([]int64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseInt(strings.TrimSpace(m), 10, 64)
		if err != nil {
			appLogger.Warnf("parse favorite dirty video_id failed: %s", m)
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	appLogger "go_zero-tiktok/pkg/logger"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// LikeCountCache 点赞计数与关系缓存，承担“关系缓冲 + 读聚合 + 脏标记”职责。
// 收藏相关方法见 favorite_count_cache.go，key 定义见 keys.go。
type LikeCountCache struct {
	rdb *redis.Redis
}

func NewLikeCountCache(rdb *redis.Redis) *LikeCountCache {
	return &LikeCountCache{rdb: rdb}
}

// LikeVideo 记录用户点赞关系，并累加计数。
// 返回 true 表示是新点赞；false 表示已经点过赞（幂等）。
func (c *LikeCountCache) LikeVideo(ctx context.Context, userID, videoID int64) (bool, error) {
	usersKey := fmtVideoLikeUsersKey(videoID)
	userVideosKey := fmtUserLikeVideosKey(userID)
	countField := strconv.FormatInt(videoID, 10)
	member := strconv.FormatInt(userID, 10)
	score := float64(time.Now().UnixMilli())

	// Lua 原子执行：加入用户集合、加入用户点赞时间线、计数+1、标记脏、续期 TTL。
	// KEYS[1]=usersKey, KEYS[2]=userVideosKey, KEYS[3]=LikeCountKey, KEYS[4]=LikeDirtyKey
	// ARGV[1]=member(user_id), ARGV[2]=score, ARGV[3]=countField, ARGV[4]=ttlSeconds
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
		[]string{usersKey, userVideosKey, LikeCountKey, LikeDirtyKey},
		[]any{member, score, countField, likeKeyTTLSeconds},
	)
	if err != nil {
		return false, xerr.Wrap(err, "LikeCountCache.LikeVideo")
	}

	v, ok := ret.(int64)
	if !ok {
		v = 0
	}
	return v == 1, nil
}

// CancelLikeVideo 移除用户点赞关系，并递减计数。
// 返回 true 表示取消成功；false 表示原本未点赞。
func (c *LikeCountCache) CancelLikeVideo(ctx context.Context, userID, videoID int64) (bool, error) {
	usersKey := fmtVideoLikeUsersKey(videoID)
	userVideosKey := fmtUserLikeVideosKey(userID)
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
		[]string{usersKey, userVideosKey, LikeCountKey, LikeDirtyKey},
		[]any{member, countField, likeKeyTTLSeconds},
	)
	if err != nil {
		return false, xerr.Wrap(err, "LikeCountCache.CancelLikeVideo")
	}

	v, ok := ret.(int64)
	if !ok {
		v = 0
	}
	return v == 1, nil
}

// IsLiked 查询用户是否点赞某视频。
func (c *LikeCountCache) IsLiked(ctx context.Context, userID, videoID int64) (bool, error) {
	usersKey := fmtVideoLikeUsersKey(videoID)
	member := strconv.FormatInt(userID, 10)
	ok, err := c.rdb.SismemberCtx(ctx, usersKey, member)
	if err != nil {
		return false, xerr.Wrap(err, "LikeCountCache.IsLiked")
	}
	return ok, nil
}

// GetLikedVideoIDs 分页获取用户点赞的视频 ID 列表（按点赞时间倒序）。
func (c *LikeCountCache) GetLikedVideoIDs(ctx context.Context, userID int64, pageNum, pageSize int32) ([]int64, int64, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	userVideosKey := fmtUserLikeVideosKey(userID)
	totalInt, err := c.rdb.ZcardCtx(ctx, userVideosKey)
	if err != nil {
		return nil, 0, xerr.Wrap(err, "LikeCountCache.GetLikedVideoIDs.Zcard")
	}
	total := int64(totalInt)

	start := int64((pageNum - 1) * pageSize)
	stop := start + int64(pageSize) - 1
	members, err := c.rdb.ZrevrangeCtx(ctx, userVideosKey, start, stop)
	if err != nil {
		return nil, 0, xerr.Wrap(err, "LikeCountCache.GetLikedVideoIDs.Zrevrange")
	}

	ids := make([]int64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseInt(strings.TrimSpace(m), 10, 64)
		if err != nil {
			appLogger.Warnf("parse liked video_id failed: %s", m)
			continue
		}
		ids = append(ids, id)
	}

	return ids, total, nil
}

// GetVideoLikeUserIDs 获取点赞某视频的全部用户 ID（供 syncer 同步关系用）。
func (c *LikeCountCache) GetVideoLikeUserIDs(ctx context.Context, videoID int64) ([]int64, error) {
	usersKey := fmtVideoLikeUsersKey(videoID)
	members, err := c.rdb.SmembersCtx(ctx, usersKey)
	if err != nil {
		return nil, xerr.Wrap(err, "LikeCountCache.GetVideoLikeUserIDs")
	}

	ids := make([]int64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseInt(strings.TrimSpace(m), 10, 64)
		if err != nil {
			appLogger.Warnf("parse like user_id failed: %s", m)
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetLikeCounts 批量获取视频 like_count。
// 返回值：命中缓存的计数、未命中的视频 ID 列表。
func (c *LikeCountCache) GetLikeCounts(ctx context.Context, videoIDs []int64) (map[int64]int64, []int64, error) {
	if len(videoIDs) == 0 {
		return map[int64]int64{}, nil, nil
	}

	fields := make([]string, 0, len(videoIDs))
	for _, id := range videoIDs {
		fields = append(fields, strconv.FormatInt(id, 10))
	}

	values, err := c.rdb.HmgetCtx(ctx, LikeCountKey, fields...)
	if err != nil {
		return nil, nil, xerr.Wrap(err, "LikeCountCache.GetLikeCounts.Hmget")
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
			appLogger.Warnf("parse like_count cache failed, video_id=%d value=%s: %v", id, values[i], err)
			missed = append(missed, id)
			continue
		}
		cached[id] = v
	}

	return cached, missed, nil
}

// SetLikeCounts 批量写入缓存（通常由 MySQL 回源后调用）。
func (c *LikeCountCache) SetLikeCounts(ctx context.Context, counts map[int64]int64) error {
	if len(counts) == 0 {
		return nil
	}

	kvs := make(map[string]string, len(counts))
	for id, count := range counts {
		kvs[strconv.FormatInt(id, 10)] = strconv.FormatInt(count, 10)
	}

	if err := c.rdb.HmsetCtx(ctx, LikeCountKey, kvs); err != nil {
		return xerr.Wrap(err, "LikeCountCache.SetLikeCounts")
	}
	// 回填时续期，避免冷 key 过期。
	_ = c.rdb.ExpireCtx(ctx, LikeCountKey, likeKeyTTLSeconds)
	return nil
}

// SetLikeCount 覆盖指定视频的 like_count 当前总值，同时移除脏标记。
func (c *LikeCountCache) SetLikeCount(ctx context.Context, videoID int64, count int64) error {
	field := strconv.FormatInt(videoID, 10)

	script := `
		redis.call('HSET', KEYS[1], ARGV[1], tonumber(ARGV[2]))
		redis.call('SREM', KEYS[2], ARGV[1])
		redis.call('EXPIRE', KEYS[1], ARGV[3])
		return 1
	`
	_, err := c.rdb.EvalCtx(ctx, script, []string{LikeCountKey, LikeDirtyKey}, []any{field, count, likeKeyTTLSeconds})
	if err != nil {
		return xerr.Wrap(err, "LikeCountCache.SetLikeCount")
	}
	return nil
}

// PopDirtyVideos 从脏集合中随机取出 batch 个待同步视频 ID（非阻塞）。
func (c *LikeCountCache) PopDirtyVideos(ctx context.Context, batch int) ([]int64, error) {
	if batch <= 0 {
		batch = 100
	}

	members, err := c.rdb.SrandmemberCtx(ctx, LikeDirtyKey, batch)
	if err != nil {
		return nil, xerr.Wrap(err, "LikeCountCache.PopDirtyVideos.Srandmember")
	}

	if len(members) == 0 {
		return nil, nil
	}

	_, err = c.rdb.SremCtx(ctx, LikeDirtyKey, toAnySlice(members)...)
	if err != nil {
		appLogger.Warnf("remove dirty videos failed: %v", err)
	}

	ids := make([]int64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseInt(strings.TrimSpace(m), 10, 64)
		if err != nil {
			appLogger.Warnf("parse dirty video_id failed: %s", m)
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

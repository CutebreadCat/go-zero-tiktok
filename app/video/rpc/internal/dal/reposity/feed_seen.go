package reposity

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// feedSeenKeyPrefix 用户曝光记录 Redis Key 前缀。
	feedSeenKeyPrefix = "feed:seen:"
)

// FeedSeenRepo 用户曝光记录仓库，基于 Redis ZSet 实现。
// member = video_id，score = 曝光时间戳(UnixMilli)，支持按时间淘汰和容量控制。
type FeedSeenRepo struct {
	rdb *redis.Redis
}

// NewFeedSeenRepo 创建曝光记录仓库。
func NewFeedSeenRepo(rdb *redis.Redis) *FeedSeenRepo {
	return &FeedSeenRepo{rdb: rdb}
}

// seenKey 构造用户曝光记录 Key。
func seenKey(userID int64) string {
	return feedSeenKeyPrefix + strconv.FormatInt(userID, 10)
}

// IsSeen 判断指定视频是否已被用户曝光。
// Redis 不可用时返回 false，降级为不去重。
func (r *FeedSeenRepo) IsSeen(ctx context.Context, userID, videoID int64) (bool, error) {
	if r.rdb == nil || userID <= 0 {
		return false, nil
	}

	_, err := r.rdb.ZscoreCtx(ctx, seenKey(userID), strconv.FormatInt(videoID, 10))
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// MarkSeen 批量标记视频为用户已曝光。
func (r *FeedSeenRepo) MarkSeen(ctx context.Context, userID int64, videoIDs []int64) error {
	if r.rdb == nil || userID <= 0 || len(videoIDs) == 0 {
		return nil
	}

	key := seenKey(userID)
	score := float64(time.Now().UnixMilli())

	// 批量写入时清理重复 video_id
	seen := make(map[int64]struct{}, len(videoIDs))
	members := make([]redis.Z, 0, len(videoIDs))
	for _, vid := range videoIDs {
		if vid <= 0 {
			continue
		}
		if _, ok := seen[vid]; ok {
			continue
		}
		seen[vid] = struct{}{}
		members = append(members, redis.Z{
			Score:  score,
			Member: strconv.FormatInt(vid, 10),
		})
	}

	if len(members) == 0 {
		return nil
	}

	return r.rdb.Pipelined(func(pipe redis.Pipeliner) error {
		for _, m := range members {
			pipe.ZAdd(ctx, key, m)
		}
		return nil
	})
}

// Cleanup 清理过期和超出容量限制的曝光记录。
// 先按 TTL 删除过期成员，再按容量删除最旧的成员。
func (r *FeedSeenRepo) Cleanup(ctx context.Context, userID int64, ttl time.Duration, maxSize int) error {
	if r.rdb == nil || userID <= 0 {
		return nil
	}

	key := seenKey(userID)

	// 1. 按 TTL 删除过期成员
	if ttl > 0 {
		cutoff := time.Now().Add(-ttl).UnixMilli()
		if _, err := r.rdb.ZremrangebyscoreCtx(ctx, key, 0, cutoff); err != nil {
			return err
		}
	}

	// 2. 按容量删除最旧的成员（score 最小的在最前面）
	if maxSize > 0 {
		count, err := r.rdb.ZcardCtx(ctx, key)
		if err != nil {
			return err
		}
		if count > maxSize {
			removeCount := int64(count - maxSize)
			if _, err := r.rdb.ZremrangebyrankCtx(ctx, key, 0, removeCount-1); err != nil {
				return err
			}
		}
	}

	return nil
}

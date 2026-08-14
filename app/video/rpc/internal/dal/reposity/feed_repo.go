package reposity

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// feedGlobalKey 全站 Feed 候选池：member=video_id, score=发布时间戳(UnixMilli)
	feedGlobalKey = "feed:global"
	// feedRetention 候选池保留窗口，超出该窗口的视频自动裁剪，退出分发。
	feedRetention = 7 * 24 * time.Hour
)

// FeedRepo 全站 Feed 候选池数据访问层。
//
// 设计原则：Redis 只存"有序 video_id 索引"，视频详情永远以 MySQL 为准（索引+水合模式）。
// 这样详情变更只需改 MySQL，索引永远是一份有序 id 列表，无双写不一致。
type FeedRepo struct {
	rdb *redis.Redis
}

// NewFeedRepo 构建候选池仓储。
func NewFeedRepo(rdb *redis.Redis) *FeedRepo {
	return &FeedRepo{rdb: rdb}
}

// AddToGlobalPool 发布成功后写入候选池，并顺手裁剪过期成员（保留最近 feedRetention）。
// 写失败仅返回错误，由调用方决定是否阻断主流程（设计上不应阻断发布）。
func (r *FeedRepo) AddToGlobalPool(ctx context.Context, videoID int64, publishAt time.Time) error {
	score := publishAt.UnixMilli()
	cutoff := time.Now().Add(-feedRetention).UnixMilli()

	// 先写入新成员，再裁剪窗口外成员（O(logN)，无需额外定时任务）
	if _, err := r.rdb.ZaddCtx(ctx, feedGlobalKey, score, strconv.FormatInt(videoID, 10)); err != nil {
		return err
	}
	_, err := r.rdb.ZremrangebyscoreCtx(ctx, feedGlobalKey, 0, cutoff)
	return err
}

// GetGlobalPoolIDs 从候选池取 (lastTimeMs, +inf] 范围内按 score 倒序的 video_id。
// lastTimeMs 为上一页最后一条的发布时间戳；传 0 表示从最新开始取。
// 返回的 id 天然按发布时间倒序排列。
func (r *FeedRepo) GetGlobalPoolIDs(ctx context.Context, lastTimeMs int64, limit int) ([]int64, error) {
	// go-zero 参数语义: start->Min(下限), stop->Max(上限), Offset=page*size, Count=size
	// ZREVRANGEBYSCORE key max min => Min=lastTimeMs+1(开区间跳过已翻过的游标), Max=+inf
	pairs, err := r.rdb.ZrevrangebyscoreWithScoresAndLimitCtx(
		ctx, feedGlobalKey, lastTimeMs+1, math.MaxInt64, 0, limit)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(pairs))
	for _, p := range pairs {
		id, err := strconv.ParseInt(p.Key, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// RemoveFromGlobalPool 从候选池移除视频（下架/删除时调用，当前项目无删除逻辑，预留）。
func (r *FeedRepo) RemoveFromGlobalPool(ctx context.Context, videoID int64) error {
	_, err := r.rdb.ZremCtx(ctx, feedGlobalKey, strconv.FormatInt(videoID, 10))
	return err
}

// PoolLen 返回候选池当前成员数（用于监控/验收）。
func (r *FeedRepo) PoolLen(ctx context.Context) (int64, error) {
	n, err := r.rdb.ZcardCtx(ctx, feedGlobalKey)
	return int64(n), err
}

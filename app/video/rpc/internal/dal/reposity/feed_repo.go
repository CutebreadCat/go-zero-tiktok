package reposity

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"go_zero-tiktok/pkg/contract" // package types
)

const (
	// feedGlobalKey 全站 Feed 候选池：member=video_id, score=发布时间戳(UnixMilli)
	feedGlobalKey = "feed:global"
	// feedRetention 候选池保留窗口，超出该窗口的视频自动裁剪，退出分发。
	feedRetention = 7 * 24 * time.Hour
)

// FeedRepo 全站 Feed 候选池 + 关注流收件箱数据访问层。
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

// inboxKey 用户关注流收件箱：member=video_id, score=发布时间戳(UnixMilli)。
// 仅存"关注的人发布的视频"，写入时机为发视频扇出（fanout on publish）。
func inboxKey(uid int64) string {
	return "feed:inbox:" + strconv.FormatInt(uid, 10)
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

// AddToUserInbox 扇出：将视频写入单个用户的收件箱 feed:inbox:{uid}。
func (r *FeedRepo) AddToUserInbox(ctx context.Context, uid, videoID int64, publishAt time.Time) error {
	_, err := r.rdb.ZaddCtx(ctx, inboxKey(uid), publishAt.UnixMilli(), strconv.FormatInt(videoID, 10))
	return err
}

// FanoutInbox 批量扇出：将视频一次性写入多个用户的收件箱（pipeline 合并 RTT）。
// 扇出属于"尽力而为"——调用方通常选择失败仅记日志、不阻断发布主流程。
func (r *FeedRepo) FanoutInbox(ctx context.Context, videoID int64, userIDs []int64, publishAt time.Time) error {
	if len(userIDs) == 0 {
		return nil
	}
	score := float64(publishAt.UnixMilli())
	member := strconv.FormatInt(videoID, 10)

	err := r.rdb.Pipelined(func(pipe redis.Pipeliner) error {
		for _, uid := range userIDs {
			pipe.ZAdd(ctx, inboxKey(uid), redis.Z{Score: score, Member: member})
		}
		return nil
	})
	return err
}

// GetGlobalPool 从候选池取 (lastTimeMs, +inf] 范围内按 score 倒序的视频索引。
// lastTimeMs 为上一页最后一条的发布时间戳；传 0 表示从最新开始取。
// 返回的索引天然按发布时间倒序排列。
func (r *FeedRepo) GetGlobalPool(ctx context.Context, lastTimeMs int64, limit int) ([]types.FeedIndex, error) {
	return r.getZSetIndexes(ctx, feedGlobalKey, lastTimeMs, limit)
}

// GetUserInbox 从用户收件箱取 (lastTimeMs, +inf] 范围内按 score 倒序的视频索引。
func (r *FeedRepo) GetUserInbox(ctx context.Context, uid, lastTimeMs int64, limit int) ([]types.FeedIndex, error) {
	return r.getZSetIndexes(ctx, inboxKey(uid), lastTimeMs, limit)
}

// getZSetIndexes 从任意 ZSet 取 (lastTimeMs, +inf] 范围按 score 倒序的 (video_id, score) 对。
func (r *FeedRepo) getZSetIndexes(ctx context.Context, key string, lastTimeMs int64, limit int) ([]types.FeedIndex, error) {
	// go-zero 参数语义: start->Min(下限), stop->Max(上限), Offset=page*size, Count=size
	// ZREVRANGEBYSCORE key max min => Min=lastTimeMs+1(开区间跳过已翻过的游标), Max=+inf
	pairs, err := r.rdb.ZrevrangebyscoreWithScoresAndLimitCtx(
		ctx, key, lastTimeMs+1, math.MaxInt64, 0, limit)
	if err != nil {
		return nil, err
	}

	indexes := make([]types.FeedIndex, 0, len(pairs))
	for _, p := range pairs {
		id, err := strconv.ParseInt(p.Key, 10, 64)
		if err != nil {
			continue
		}
		indexes = append(indexes, types.FeedIndex{VideoID: id, Score: int64(p.Score)})
	}
	return indexes, nil
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

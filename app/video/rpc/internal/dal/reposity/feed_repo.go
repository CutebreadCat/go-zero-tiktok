package reposity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math"
	"strconv"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"go_zero-tiktok/pkg/contract" // package types
)

const (
	// feedGlobalKey 全站 Feed 候选池：member=video_id, score=发布时间戳(UnixMilli)
	feedGlobalKey = "feed:global"
	// feedRetention 候选池保留窗口，超出该窗口的视频自动裁剪，退出分发。
	feedRetention = 7 * 24 * time.Hour
	// hotVideosKey 全站热门视频有序集合：member=video_id, score=热度分（放大 1000 倍）
	hotVideosKey = "hot:videos"
	// hotRebuildKey 热榜重建临时键：构建完成后原子 RENAME 到 hotVideosKey。
	// 注意：RENAME 要求两个 key 同槽，当前部署为单节点 Redis；若切集群需给两键加同一 hash tag。
	hotRebuildKey = "hot:videos:rebuild"
	// hotRebuildLockKey 重建互斥锁，value 为本次执行的唯一 token。
	hotRebuildLockKey = "hot:videos:rebuild:lock"
	// hotFetchFactor Redis 同分过滤兜底因子，先多取若干条再按 video_id 精断
	hotFetchFactor = 5
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

// RefreshHotScore 覆盖更新视频热度分。
func (r *FeedRepo) RefreshHotScore(ctx context.Context, videoID int64, score int64) error {
	if videoID == 0 {
		return nil
	}
	member := strconv.FormatInt(videoID, 10)
	if _, err := r.rdb.ZaddCtx(ctx, hotVideosKey, score, member); err != nil {
		return pkgerrors.WithMessage(err, "FeedRepo.RefreshHotScore")
	}
	return nil
}

// RefreshHotScores 批量覆盖正式榜单分数。
// 只更新已存在成员（ZADD 不会因为覆盖而丢失其它成员），分块 pipeline 降低 RTT。
func (r *FeedRepo) RefreshHotScores(ctx context.Context, scores []types.FeedIndex) error {
	if len(scores) == 0 {
		return nil
	}

	const batchSize = 500
	for start := 0; start < len(scores); start += batchSize {
		end := start + batchSize
		if end > len(scores) {
			end = len(scores)
		}
		part := scores[start:end]

		err := r.rdb.PipelinedCtx(ctx, func(pipe redis.Pipeliner) error {
			for _, item := range part {
				if item.VideoID <= 0 {
					continue
				}
				pipe.ZAdd(ctx, hotVideosKey, redis.Z{
					Score:  float64(item.Score),
					Member: strconv.FormatInt(item.VideoID, 10),
				})
			}
			return nil
		})
		if err != nil {
			return pkgerrors.WithMessage(err, "FeedRepo.RefreshHotScores")
		}
	}
	return nil
}

// GetHotScores 批量读取视频当前热度分（未上榜或读取失败按 0 返回）。
// 用 pipeline 一次拿完，避免按 id 逐个往返。
func (r *FeedRepo) GetHotScores(ctx context.Context, videoIDs []int64) (map[int64]int64, error) {
	scores := make(map[int64]int64, len(videoIDs))
	if len(videoIDs) == 0 {
		return scores, nil
	}

	// 用接口接收 *redis.FloatCmd，避免直接依赖 go-redis 的具体命令类型。
	type floatResult interface {
		Result() (float64, error)
	}
	cmds := make([]floatResult, len(videoIDs))
	err := r.rdb.PipelinedCtx(ctx, func(pipe redis.Pipeliner) error {
		for i, id := range videoIDs {
			cmds[i] = pipe.ZScore(ctx, hotVideosKey, strconv.FormatInt(id, 10))
		}
		return nil
	})
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "FeedRepo.GetHotScores")
	}

	for i, id := range videoIDs {
		score, err := cmds[i].Result()
		if err != nil {
			continue // 未上榜（redis.Nil）或单条失败，按 0 处理
		}
		scores[id] = int64(score)
	}
	return scores, nil
}

// RemoveHotVideos 从热门榜单移除指定成员（详情已不存在时调用），避免榜单出现空洞。
func (r *FeedRepo) RemoveHotVideos(ctx context.Context, videoIDs []int64) error {
	if len(videoIDs) == 0 {
		return nil
	}
	members := make([]any, 0, len(videoIDs))
	for _, id := range videoIDs {
		if id <= 0 {
			continue
		}
		members = append(members, strconv.FormatInt(id, 10))
	}
	if len(members) == 0 {
		return nil
	}
	if _, err := r.rdb.ZremCtx(ctx, hotVideosKey, members...); err != nil {
		return pkgerrors.WithMessage(err, "FeedRepo.RemoveHotVideos")
	}
	return nil
}

// TrimHotVideos 只保留热度最高的 keepTopN 个成员。
// 榜单退出由排名决定，不再按"最后活跃时间"淘汰：
// 否则低流量或服务停摆超过窗口时长时，整个榜单会被清理器一次性清空。
func (r *FeedRepo) TrimHotVideos(ctx context.Context, keepTopN int) error {
	if keepTopN <= 0 {
		return nil
	}
	// ZREMRANGEBYRANK 按 score 升序排名，-1 表示分数最高的成员，
	// 因此保留最高分的 keepTopN 个等价于删除 [0, -(keepTopN+1)]。
	if _, err := r.rdb.ZremrangebyrankCtx(ctx, hotVideosKey, 0, -int64(keepTopN)-1); err != nil {
		return pkgerrors.WithMessage(err, "FeedRepo.TrimHotVideos")
	}
	return nil
}

// GetHotVideosByCursor 从 hot:videos 取热度分 <= cursorScore 的索引，按热度倒序。
// Redis ZSet 只能按 score 范围查询，同分场景下多取 hotFetchFactor*limit 条，
// 由调用方按 (score, video_id) < (cursorScore, cursorVideoID) 精断后截断。
//
// 注意 go-zero 的 ZrevrangebyscoreWithScoresAndLimitCtx 参数语义是
// (key, start=Min, stop=Max)，对应 ZREVRANGEBYSCORE key Max Min，不能写反。
func (r *FeedRepo) GetHotVideosByCursor(ctx context.Context, cursorScore, cursorVideoID int64, limit int) ([]types.FeedIndex, error) {
	if limit <= 0 {
		return nil, nil
	}

	// 首页（cursorScore==0）从最高分开始取，上界为 +inf；
	// 翻页时先取严格小于游标分的部分，同分成员在下面按 video_id 补。
	maxScore := cursorScore - 1
	if cursorScore == 0 {
		maxScore = math.MaxInt64
	}
	pairs, err := r.rdb.ZrevrangebyscoreWithScoresAndLimitCtx(
		ctx, hotVideosKey, math.MinInt64, maxScore, 0, limit)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "FeedRepo.GetHotVideosByCursor")
	}

	// 首页没有同分游标需要精断，直接返回
	if cursorScore == 0 {
		return r.pairsToIndexes(pairs), nil
	}

	// 取同分部分，客户端过滤
	tiePairs, err := r.rdb.ZrevrangebyscoreWithScoresAndLimitCtx(
		ctx, hotVideosKey, cursorScore, cursorScore, 0, limit*hotFetchFactor)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "FeedRepo.GetHotVideosByCursor")
	}
	for _, p := range tiePairs {
		id, err := strconv.ParseInt(p.Key, 10, 64)
		if err != nil {
			continue
		}
		if id < cursorVideoID {
			pairs = append(pairs, p)
		}
	}

	return r.pairsToIndexes(pairs), nil
}

// ResetHotRebuild 清理上次残留的重建临时键。
func (r *FeedRepo) ResetHotRebuild(ctx context.Context) error {
	if _, err := r.rdb.DelCtx(ctx, hotRebuildKey); err != nil {
		return pkgerrors.WithMessage(err, "FeedRepo.ResetHotRebuild")
	}
	return nil
}

// AppendHotRebuildScores 分批把分值写入重建临时键（ZADD 覆盖写）。
func (r *FeedRepo) AppendHotRebuildScores(ctx context.Context, scores []types.FeedIndex) error {
	if len(scores) == 0 {
		return nil
	}

	const batchSize = 500
	for start := 0; start < len(scores); start += batchSize {
		end := start + batchSize
		if end > len(scores) {
			end = len(scores)
		}
		part := scores[start:end]

		err := r.rdb.PipelinedCtx(ctx, func(pipe redis.Pipeliner) error {
			for _, item := range part {
				if item.VideoID <= 0 {
					continue
				}
				pipe.ZAdd(ctx, hotRebuildKey, redis.Z{
					Score:  float64(item.Score),
					Member: strconv.FormatInt(item.VideoID, 10),
				})
			}
			return nil
		})
		if err != nil {
			return pkgerrors.WithMessage(err, "FeedRepo.AppendHotRebuildScores")
		}
	}

	return nil
}

// PublishHotRebuild 用临时键原子替换正式榜单，并按 keepTopN 裁剪，返回最终成员数。
// 替换用 Lua 完成：读请求要么看到旧榜、要么看到新榜，不会读到半成品。
func (r *FeedRepo) PublishHotRebuild(ctx context.Context, keepTopN int) (int64, error) {
	const script = `
if redis.call('EXISTS', KEYS[1]) == 0 then
  return 0
end
redis.call('RENAME', KEYS[1], KEYS[2])
local keep = tonumber(ARGV[1])
if keep and keep > 0 then
  redis.call('ZREMRANGEBYRANK', KEYS[2], 0, -(keep + 1))
end
return redis.call('ZCARD', KEYS[2])`

	res, err := r.rdb.EvalCtx(ctx, script, []string{hotRebuildKey, hotVideosKey}, keepTopN)
	if err != nil {
		return 0, pkgerrors.WithMessage(err, "FeedRepo.PublishHotRebuild")
	}

	count, _ := res.(int64)
	return count, nil
}

// TryLockHotRebuild 抢占重建锁，返回本次执行的 token 与是否成功。
func (r *FeedRepo) TryLockHotRebuild(ctx context.Context, ttl time.Duration) (string, bool, error) {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", false, pkgerrors.WithMessage(err, "FeedRepo.TryLockHotRebuild")
	}
	token := hex.EncodeToString(buf)

	ok, err := r.rdb.SetnxExCtx(ctx, hotRebuildLockKey, token, int(ttl.Seconds()))
	if err != nil {
		return "", false, pkgerrors.WithMessage(err, "FeedRepo.TryLockHotRebuild")
	}
	return token, ok, nil
}

// UnlockHotRebuild 仅在 token 与锁值一致时释放，避免 TTL 过期后误删下一个执行者的锁。
func (r *FeedRepo) UnlockHotRebuild(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	const script = `
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0`

	if _, err := r.rdb.EvalCtx(ctx, script, []string{hotRebuildLockKey}, token); err != nil {
		return pkgerrors.WithMessage(err, "FeedRepo.UnlockHotRebuild")
	}
	return nil
}

// HotRebuildRunning 探测重建锁是否存在，供分数保鲜任务避让正在进行的全量重建。
func (r *FeedRepo) HotRebuildRunning(ctx context.Context) (bool, error) {
	running, err := r.rdb.ExistsCtx(ctx, hotRebuildLockKey)
	if err != nil {
		return false, pkgerrors.WithMessage(err, "FeedRepo.HotRebuildRunning")
	}
	return running, nil
}

func (r *FeedRepo) pairsToIndexes(pairs []redis.Pair) []types.FeedIndex {
	indexes := make([]types.FeedIndex, 0, len(pairs))
	for _, p := range pairs {
		id, err := strconv.ParseInt(p.Key, 10, 64)
		if err != nil {
			continue
		}
		indexes = append(indexes, types.FeedIndex{VideoID: id, Score: int64(p.Score)})
	}
	return indexes
}

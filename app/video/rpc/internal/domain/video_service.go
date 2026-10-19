package domain

import (
	"context"
	"fmt"
	"time"

	feedpkg "go_zero-tiktok/app/video/rpc/internal/domain/feed"
	"go_zero-tiktok/pkg/contract"

	"github.com/zeromicro/go-zero/core/logx"
)

// VisitEventProducer 视频访问事件生产者接口，用于异步化访问量刷新。
type VisitEventProducer interface {
	Send(ctx context.Context, videoID int64, delta int64) error
	Close() error
}

// VideoService 视频领域服务，聚焦视频发布、Feed、搜索、热度排序等核心能力。
// 点赞/收藏等互动能力已拆分到 interaction 子领域。
type VideoService struct {
	videoRepo          IVideoRepo
	popularRepo        IPopularRepo
	qosRepo            IVideoQoSRepo
	storage            StorageProvider
	feedRepo           IFeedRepo
	hotScoreCalculator *feedpkg.HotScoreCalculator
	strategyFactory    *feedpkg.StrategyFactory
	visitProducer      VisitEventProducer
}

func NewVideoService(
	videoRepo IVideoRepo,
	popularRepo IPopularRepo,
	qosRepo IVideoQoSRepo,
	storage StorageProvider,
	feedRepo IFeedRepo,
	seenRepo ISeenRepo,
	hotScoreCalculator *feedpkg.HotScoreCalculator,
	visitProducer VisitEventProducer,
	recommendConfig feedpkg.RecommendConfig,
) *VideoService {
	return &VideoService{
		videoRepo:          videoRepo,
		popularRepo:        popularRepo,
		qosRepo:            qosRepo,
		storage:            storage,
		feedRepo:           feedRepo,
		visitProducer:      visitProducer,
		hotScoreCalculator: hotScoreCalculator,
		strategyFactory: feedpkg.NewStrategyFactory(
			feedpkg.NewTimelineStrategy(videoRepo, popularRepo, feedRepo),
			feedpkg.NewFollowingStrategy(videoRepo, popularRepo, feedRepo),
			feedpkg.NewHotStrategy(videoRepo, feedRepo),
			feedpkg.NewRecommendStrategy(videoRepo, popularRepo, feedRepo, qosRepo, seenRepo, recommendConfig),
		),
	}
}

// PublishVideo 创建视频并初始化统计记录，发布成功后写入 Feed 候选池与热门榜。
// idempotencyKey 非空时按 (authorID, key) 幂等：重复/并发提交复用已存在的 videoID，不会重复发布。
// 返回最终生效的 videoID 与发布时刻 publishAt（幂等命中时为首次发布时间）。
// 候选池写入失败不阻断发布（DB 是主数据源，Redis 是加速层）。
func (s *VideoService) PublishVideo(ctx context.Context, videoID, authorID int64, videoObjectKey, coverObjectKey, title, description, idempotencyKey string) (int64, time.Time, error) {
	publishAt := time.Now()
	actualID, err := s.videoRepo.CreateVideoFromParams(ctx, videoID, authorID, videoObjectKey, coverObjectKey, title, description, idempotencyKey)
	if err != nil {
		return 0, time.Time{}, err
	}

	// 并发重试：本次生成的新 ID 未生效，复用先写入记录及其原始发布时间
	if actualID != videoID {
		publishAt, err = s.videoRepo.GetVideoPublishAt(ctx, actualID)
		if err != nil {
			return 0, time.Time{}, err
		}
	}

	if err := s.ensureVideoIndexed(ctx, actualID, publishAt); err != nil {
		return 0, time.Time{}, err
	}

	return actualID, publishAt, nil
}

// FindPublishedVideoByKey 幂等键命中时返回已发布视频，并补齐可能缺失的统计行/Feed 索引。
// 这样"首次发布部分失败 + 客户端重试"也能收敛到完整状态，而不是直接返回半成品。
func (s *VideoService) FindPublishedVideoByKey(ctx context.Context, authorID int64, idempotencyKey string) (int64, time.Time, bool, error) {
	if idempotencyKey == "" {
		return 0, time.Time{}, false, nil
	}

	videoID, publishAt, found, err := s.videoRepo.GetVideoByAuthorAndIdempotencyKey(ctx, authorID, idempotencyKey)
	if err != nil || !found {
		return 0, time.Time{}, found, err
	}

	if err := s.ensureVideoIndexed(ctx, videoID, publishAt); err != nil {
		return 0, time.Time{}, false, err
	}

	return videoID, publishAt, true, nil
}

// ensureVideoIndexed 初始化视频的统计行与 Feed/热度索引。
// 全部操作幂等（统计行 INSERT IGNORE、ZSet 覆盖写），可安全重复执行。
func (s *VideoService) ensureVideoIndexed(ctx context.Context, videoID int64, publishAt time.Time) error {
	if err := s.popularRepo.CreatePopularVideo(ctx, videoID); err != nil {
		return err
	}

	// 写入候选池 feed:global（失败仅记日志，不影响发布主流程）
	if s.feedRepo != nil {
		if err := s.feedRepo.AddToGlobalPool(ctx, videoID, publishAt); err != nil {
			logx.Errorf("add video %d to feed global pool failed: %v", videoID, err)
		}
	}

	// 同步计算初始热度分并写入 hot:videos（Redis 写入失败不阻断发布）。
	if err := s.RecalculateHotScore(ctx, videoID); err != nil {
		logx.Errorf("recalculate hot score for new video %d failed: %v", videoID, err)
	}

	return nil
}

// FanoutToUsers 关注流扇出：将视频写入每个粉丝的收件箱 feed:inbox:{uid}。
// 扇出属于"尽力而为"——失败仅记日志返回错误，由调用方决定是否阻断发布（设计上不应阻断）。
// 粉丝列表由 gateway 编排层从 communication.rpc 获取，本方法只写 video.rpc 自己的 Redis 索引。
func (s *VideoService) FanoutToUsers(ctx context.Context, videoID int64, userIDs []int64, publishAt time.Time) error {
	if s.feedRepo == nil || len(userIDs) == 0 {
		return nil
	}
	if err := s.feedRepo.FanoutInbox(ctx, videoID, userIDs, publishAt); err != nil {
		logx.Errorf("fanout video %d to %d users failed: %v", videoID, len(userIDs), err)
		return err
	}
	return nil
}

// IncreaseVideoVisitCount 同步增加视频访问量，并触发热度分重算。
// 热度分写入 Redis 失败不阻断访问计数主流程。
func (s *VideoService) IncreaseVideoVisitCount(ctx context.Context, videoID int64, delta int64) error {
	if err := s.popularRepo.IncreaseVideoVisitCount(ctx, videoID, delta); err != nil {
		return err
	}
	if err := s.RecalculateHotScore(ctx, videoID); err != nil {
		logx.Errorf("recalculate hot score failed for video %d: %v", videoID, err)
	}
	return nil
}

// RecalculateHotScore 根据 MySQL 统计与发布时间重新计算并覆盖 Redis 热度分。
func (s *VideoService) RecalculateHotScore(ctx context.Context, videoID int64) error {
	if videoID == 0 || s.feedRepo == nil || s.hotScoreCalculator == nil {
		return nil
	}

	publishAt, err := s.videoRepo.GetVideoPublishAt(ctx, videoID)
	if err != nil {
		return err
	}

	populars, err := s.popularRepo.GetPopularVideosByIDs(ctx, []int64{videoID})
	if err != nil {
		return err
	}
	popular, ok := populars[videoID]
	if !ok {
		popular = types.VideoPopular{VideoID: videoID}
	}

	score := s.hotScoreCalculator.Compute(popular, publishAt)
	return s.feedRepo.RefreshHotScore(ctx, videoID, score)
}

// hotRebuildDefaultBatch 重建时单批扫描的统计行数。
const hotRebuildDefaultBatch int32 = 500

// HotRebuildOptions 热度榜重建参数。
type HotRebuildOptions struct {
	Since     time.Time // 候选窗口起点；零值表示全量重建
	BatchSize int32     // 单批扫描行数，<=0 取默认 500
	KeepTopN  int       // 最终保留条数，<=0 不裁剪
	DryRun    bool      // 只统计候选数量，不写 Redis
}

// RebuildHotRanking 从 MySQL 计数重建 hot:videos。
//
//  1. 按候选窗口分批扫描（窗口模式走 (created_at, video_id) 复合索引）；
//  2. 用与实时重算相同的公式算分，写入临时键；
//  3. Lua 原子 RENAME 到正式键并按 KeepTopN 裁剪。
//
// 返回扫描条数与最终榜单成员数；DryRun 时只返回候选数，members 为 0。
// 重建期间读请求始终读旧榜，失败不会替换正式榜。
func (s *VideoService) RebuildHotRanking(ctx context.Context, opts HotRebuildOptions) (scanned, members int64, err error) {
	if s.feedRepo == nil || s.popularRepo == nil || s.hotScoreCalculator == nil {
		return 0, 0, fmt.Errorf("hot ranking rebuild dependencies are not configured")
	}
	if opts.BatchSize <= 0 {
		opts.BatchSize = hotRebuildDefaultBatch
	}

	// dry-run 只读，不抢锁、不写 Redis。
	if opts.DryRun {
		scanned, err = s.popularRepo.CountVideoStatsForRebuild(ctx, opts.Since)
		return scanned, 0, err
	}

	token, locked, err := s.feedRepo.TryLockHotRebuild(ctx, 30*time.Minute)
	if err != nil {
		return 0, 0, err
	}
	if !locked {
		return 0, 0, fmt.Errorf("hot ranking rebuild is already running")
	}

	published := false
	defer func() {
		cleanupCtx := context.WithoutCancel(ctx)
		// 成功时临时键已被 RENAME 消费；失败时清理，避免残留半成品。
		if !published {
			if cleanupErr := s.feedRepo.ResetHotRebuild(cleanupCtx); cleanupErr != nil {
				logx.Errorf("clean hot ranking rebuild temp key failed: %v", cleanupErr)
			}
		}
		if unlockErr := s.feedRepo.UnlockHotRebuild(cleanupCtx, token); unlockErr != nil {
			logx.Errorf("release hot ranking rebuild lock failed: %v", unlockErr)
		}
	}()

	if err = s.feedRepo.ResetHotRebuild(ctx); err != nil {
		return 0, 0, err
	}

	var cursorAt time.Time
	var cursorID int64
	if !opts.Since.IsZero() {
		// 窗口模式首页游标为 (cutoff, 0)，包含 created_at 恰好等于 cutoff 的视频。
		cursorAt = opts.Since
	}

	for {
		rows, listErr := s.popularRepo.ListVideoStatsForRebuild(ctx, opts.Since, cursorAt, cursorID, opts.BatchSize)
		if listErr != nil {
			return 0, 0, listErr
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]types.FeedIndex, 0, len(rows))
		for _, row := range rows {
			popular := types.VideoPopular{
				VideoID:       row.VideoID,
				VisitCount:    row.VisitCount,
				LikeCount:     row.LikeCount,
				CommentCount:  row.CommentCount,
				FavoriteCount: row.FavoriteCount,
			}
			batch = append(batch, types.FeedIndex{
				VideoID: row.VideoID,
				Score:   s.hotScoreCalculator.Compute(popular, row.PublishAt),
			})
			cursorAt = row.PublishAt
			cursorID = row.VideoID
		}

		if appendErr := s.feedRepo.AppendHotRebuildScores(ctx, batch); appendErr != nil {
			return 0, 0, appendErr
		}
		scanned += int64(len(batch))
		logx.Infof("hot ranking rebuild progress: scanned=%d cursor_at=%s cursor_id=%d",
			scanned, cursorAt.Format(time.RFC3339Nano), cursorID)
	}

	// 空结果不替换正式榜，避免异常窗口或查询故障把线上榜单清空。
	if scanned == 0 {
		return 0, 0, fmt.Errorf("no video stats found, abort hot ranking rebuild")
	}

	members, err = s.feedRepo.PublishHotRebuild(ctx, opts.KeepTopN)
	if err != nil {
		return scanned, 0, err
	}
	published = true

	logx.Infof("hot ranking rebuild done: scanned=%d members=%d keep_top_n=%d", scanned, members, opts.KeepTopN)
	return scanned, members, nil
}

// RefreshHotRankingScores 按当前时间重算榜单前 limit 名成员的热度分。
//
// 热度分是写入时刻的快照，视频停止互动后分数不会自然衰减；定时调用本方法
// 可以让榜单持续逼近 HN 公式的真实值。只覆盖已上榜成员，不新增成员，
// 新视频仍由实时重算链路写入。
func (s *VideoService) RefreshHotRankingScores(ctx context.Context, limit int) (int, error) {
	if limit <= 0 || s.feedRepo == nil || s.popularRepo == nil || s.hotScoreCalculator == nil {
		return 0, nil
	}

	indexes, err := s.feedRepo.GetHotVideosByCursor(ctx, 0, 0, limit)
	if err != nil {
		return 0, err
	}
	if len(indexes) == 0 {
		return 0, nil
	}

	videoIDs := make([]int64, 0, len(indexes))
	for _, idx := range indexes {
		videoIDs = append(videoIDs, idx.VideoID)
	}

	rows, err := s.popularRepo.ListVideoStatsByIDs(ctx, videoIDs)
	if err != nil {
		return 0, err
	}

	batch := make([]types.FeedIndex, 0, len(rows))
	existing := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		popular := types.VideoPopular{
			VideoID:       row.VideoID,
			VisitCount:    row.VisitCount,
			LikeCount:     row.LikeCount,
			CommentCount:  row.CommentCount,
			FavoriteCount: row.FavoriteCount,
		}
		batch = append(batch, types.FeedIndex{
			VideoID: row.VideoID,
			Score:   s.hotScoreCalculator.Compute(popular, row.PublishAt),
		})
		existing[row.VideoID] = struct{}{}
	}

	if err := s.feedRepo.RefreshHotScores(ctx, batch); err != nil {
		return 0, err
	}

	// 榜内但 MySQL 已查不到的成员（视频删除/统计行缺失）顺手摘除，避免榜单空洞。
	s.removeStaleHotIndexes(ctx, videoIDs, existing)

	logx.Infof("hot ranking scores refreshed: candidates=%d refreshed=%d", len(videoIDs), len(batch))
	return len(batch), nil
}

// removeStaleHotIndexes 把榜单里查不到统计行的成员移除（视频已删除或数据缺失）。
func (s *VideoService) removeStaleHotIndexes(ctx context.Context, requested []int64, existing map[int64]struct{}) {
	stale := make([]int64, 0)
	for _, id := range requested {
		if _, ok := existing[id]; !ok {
			stale = append(stale, id)
		}
	}
	if len(stale) == 0 {
		return
	}

	logx.Infof("removing %d stale videos from hot ranking", len(stale))
	if err := s.feedRepo.RemoveHotVideos(ctx, stale); err != nil {
		logx.Errorf("remove stale hot videos failed: %v", err)
	}
}

// RecordVisit 记录视频访问量与热度分。
// 若配置了访问事件生产者，则通过 Kafka 异步化（业务代码不创建 goroutine）；
// 否则同步执行 IncreaseVideoVisitCount 作为兜底。
func (s *VideoService) RecordVisit(ctx context.Context, videoID int64) {
	if videoID == 0 {
		return
	}
	if s.visitProducer != nil {
		if err := s.visitProducer.Send(ctx, videoID, 1); err != nil {
			logx.Errorf("send video visit event failed, video_id=%d: %v", videoID, err)
		}
		return
	}
	if err := s.IncreaseVideoVisitCount(ctx, videoID, 1); err != nil {
		logx.Errorf("record visit failed for video %d: %v", videoID, err)
	}
}

// recordVisits 批量记录访问量。
func (s *VideoService) recordVisits(ctx context.Context, videos []types.VideoBaseinfo) {
	for _, video := range videos {
		s.RecordVisit(ctx, video.VideoID)
	}
}

// GetVideosByIDs 根据视频 ID 列表批量查询视频信息。
func (s *VideoService) GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error) {
	return s.videoRepo.GetVideosByIDs(ctx, videoIDs)
}

// GetPopularVideos 获取热门视频列表（ID 水合为完整视频信息），并拼接 QoS 聚合指标。
func (s *VideoService) GetPopularVideos(ctx context.Context, pageNum, pageSize int32) ([]types.VideoBaseinfo, []types.VideoPopular, error) {
	videoPopulars, _, err := s.popularRepo.GetPopularVideoIDsByVisitCount(ctx, pageNum, pageSize)
	if err != nil {
		return nil, nil, err
	}

	videoIDs := make([]int64, 0, len(videoPopulars))
	for _, vp := range videoPopulars {
		videoIDs = append(videoIDs, vp.VideoID)
	}

	videos, err := s.videoRepo.GetVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, nil, err
	}

	// 合并 QoS 指标（失败仅记日志，不影响 popular 主链路）。
	if s.qosRepo != nil {
		if err := s.mergeQoSMetrics(ctx, videoPopulars); err != nil {
			logx.Errorf("merge qos metrics failed: %v", err)
		}
	}

	return videos, videoPopulars, nil
}

// mergeQoSMetrics 把 video_qos_stat 指标合并到 VideoPopular 列表。
func (s *VideoService) mergeQoSMetrics(ctx context.Context, populars []types.VideoPopular) error {
	if len(populars) == 0 {
		return nil
	}
	videoIDs := make([]int64, 0, len(populars))
	for _, p := range populars {
		videoIDs = append(videoIDs, p.VideoID)
	}

	metrics, err := s.qosRepo.GetQoSMetricsByVideoIDs(ctx, videoIDs)
	if err != nil {
		return err
	}

	for i := range populars {
		populars[i].VideoQoSMetrics = metrics[populars[i].VideoID]
	}
	return nil
}

// SearchVideos 搜索视频并记录访问量
func (s *VideoService) SearchVideos(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := s.videoRepo.SearchVideosByKeyword(ctx, keyword, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.recordVisits(ctx, videos)
	return videos, total, nil
}

// GetVideosByAuthor 获取作者视频列表并记录访问量
func (s *VideoService) GetVideosByAuthor(ctx context.Context, authorID int64, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := s.videoRepo.GetVideosByAuthorID(ctx, authorID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.recordVisits(ctx, videos)
	return videos, total, nil
}

// GetFeedVideos 获取 Feed 流。viewerID 为当前浏览用户（<=0 表示无登录态，仅看全站候选池）。
// 根据 scene 路由到对应策略，默认 scene 为 timeline。
// 返回视频基础信息、热度统计、下一页游标、是否还有更多。
func (s *VideoService) GetFeedVideos(ctx context.Context, viewerID int64, scene, cursor string, limit int32) (*feedpkg.Result, error) {
	strategy, err := s.strategyFactory.Get(scene)
	if err != nil {
		return nil, err
	}

	result, err := strategy.GetFeed(ctx, viewerID, cursor, limit)
	if err != nil {
		return nil, err
	}

	s.recordVisits(ctx, result.Videos)
	return result, nil
}

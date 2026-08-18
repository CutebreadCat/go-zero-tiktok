package feed

import (
	"context"
	"strings"
	"time"

	"go_zero-tiktok/pkg/contract"
	myutils "go_zero-tiktok/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// TimelineStrategy 时间线策略：基于 feed:global 候选池与 MySQL 兜底，按发布时间倒序返回。
// 为兼容旧接口，有登录态时会合并 feed:inbox:{uid} 关注收件箱。
type TimelineStrategy struct {
	videoRepo   domainVideoRepo
	popularRepo domainPopularRepo
	feedRepo    domainFeedRepo
}

// domainVideoRepo 是 TimelineStrategy 所需的视频仓储能力子集。
type domainVideoRepo interface {
	GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error)
	GetVideoByLastTime(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error)
}

// domainPopularRepo 是 TimelineStrategy 所需的热度仓储能力子集。
type domainPopularRepo interface {
	GetPopularVideosByIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoPopular, error)
}

// domainFeedRepo 是 TimelineStrategy 所需的 Feed 索引仓储能力子集。
type domainFeedRepo interface {
	GetGlobalPool(ctx context.Context, lastTimeMs int64, limit int) ([]types.FeedIndex, error)
	GetUserInbox(ctx context.Context, uid, lastTimeMs int64, limit int) ([]types.FeedIndex, error)
}

// NewTimelineStrategy 创建时间线策略。
func NewTimelineStrategy(videoRepo domainVideoRepo, popularRepo domainPopularRepo, feedRepo domainFeedRepo) *TimelineStrategy {
	return &TimelineStrategy{
		videoRepo:   videoRepo,
		popularRepo: popularRepo,
		feedRepo:    feedRepo,
	}
}

// Name 返回策略名。
func (s *TimelineStrategy) Name() string {
	return "timeline"
}

// GetFeed 读取时间线 Feed 一页。
func (s *TimelineStrategy) GetFeed(ctx context.Context, viewerID int64, cursor string, limit int32) (*Result, error) {
	parsedCursor, err := DecodeTimelineCursor(cursor)
	if err != nil {
		return nil, err
	}

	videos, populars, hasMore, err := s.getFeedFromPools(ctx, viewerID, parsedCursor, limit)
	if err != nil {
		return nil, err
	}

	total := int64(len(videos))
	if hasMore {
		total = int64(len(videos)) + 1 // 兼容旧 total 语义：还有下一页时多 1
	}

	var nextCursor string
	if len(videos) > 0 {
		last := videos[len(videos)-1]
		lastMs, err := lastPublishedAtMs(last.CreatedAt)
		if err != nil {
			logx.Errorf("parse last video created_at failed: %v", err)
		} else {
			nextCursor = EncodeTimelineCursor(&TimelineCursor{
				PublishedAt: lastMs,
				VideoID:     last.VideoID,
			})
		}
	}

	return &Result{
		Videos:     videos,
		Populars:   populars,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      total,
	}, nil
}

// getFeedFromPools 从 Redis 索引取视频，候选池不足时走 MySQL 兜底。
// 有登录态时合并 feed:global 与 feed:inbox:{viewerID}。
func (s *TimelineStrategy) getFeedFromPools(ctx context.Context, viewerID int64, cursor *TimelineCursor, pageSize int32) ([]types.VideoBaseinfo, []types.VideoPopular, bool, error) {
	var lastTimeMs int64
	var lastTimeStr string
	if cursor != nil {
		lastTimeMs = cursor.PublishedAt
		lastTimeStr = myutils.TimeToStr(time.UnixMilli(lastTimeMs), "")
	}

	if s.feedRepo != nil {
		videos, populars, hasMore, err := s.getFeedFromRedis(ctx, viewerID, lastTimeMs, lastTimeStr, pageSize)
		if err != nil {
			logx.Errorf("get feed from redis failed, fallback to db: %v", err)
		} else if len(videos) > 0 {
			return videos, populars, hasMore, nil
		}
	}

	// 兜底：MySQL 直查
	videos, _, err := s.videoRepo.GetVideoByLastTime(ctx, lastTimeStr, 1, pageSize)
	if err != nil {
		return nil, nil, false, err
	}
	populars := s.batchGetPopulars(ctx, videos)
	return videos, populars, false, nil
}

// getFeedFromRedis 从 Redis 候选池读取并合并。
func (s *TimelineStrategy) getFeedFromRedis(ctx context.Context, viewerID int64, lastTimeMs int64, lastTimeStr string, pageSize int32) ([]types.VideoBaseinfo, []types.VideoPopular, bool, error) {
	global, err := s.feedRepo.GetGlobalPool(ctx, lastTimeMs, int(pageSize)+1)
	if err != nil {
		return nil, nil, false, err
	}

	var inbox []types.FeedIndex
	if viewerID > 0 {
		inbox, err = s.feedRepo.GetUserInbox(ctx, viewerID, lastTimeMs, int(pageSize)+1)
		if err != nil {
			return nil, nil, false, err
		}
	}

	merged, hasMore := mergeFeedIndexes(global, inbox, pageSize)
	if len(merged) == 0 {
		return nil, nil, false, nil
	}

	videoIDs := make([]int64, 0, len(merged))
	for _, fi := range merged {
		videoIDs = append(videoIDs, fi.VideoID)
	}

	videos, err := s.videoRepo.GetVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, nil, false, err
	}

	populars := s.batchGetPopulars(ctx, videos)
	return videos, populars, hasMore, nil
}

// mergeFeedIndexes 将两条按 score 倒序的流按 score 归并、去重，截取前 limit 条。
func mergeFeedIndexes(global, inbox []types.FeedIndex, limit int32) ([]types.FeedIndex, bool) {
	cap := len(global) + len(inbox)
	merged := make([]types.FeedIndex, 0, cap)
	i, j := 0, 0
	for i < len(global) && j < len(inbox) {
		a, b := global[i], inbox[j]
		switch {
		case a.VideoID == b.VideoID:
			merged = append(merged, a)
			i++
			j++
		case a.Score > b.Score:
			merged = append(merged, a)
			i++
		default:
			merged = append(merged, b)
			j++
		}
		if len(merged) > int(limit) {
			return merged[:limit], true
		}
	}
	for ; i < len(global); i++ {
		merged = append(merged, global[i])
		if len(merged) > int(limit) {
			return merged[:limit], true
		}
	}
	for ; j < len(inbox); j++ {
		merged = append(merged, inbox[j])
		if len(merged) > int(limit) {
			return merged[:limit], true
		}
	}

	return merged, false
}

// batchGetPopulars 根据视频列表批量查询 visit_count 等热度统计。
func (s *TimelineStrategy) batchGetPopulars(ctx context.Context, videos []types.VideoBaseinfo) []types.VideoPopular {
	if len(videos) == 0 {
		return nil
	}

	videoIDs := make([]int64, 0, len(videos))
	for _, v := range videos {
		videoIDs = append(videoIDs, v.VideoID)
	}

	populars, err := s.popularRepo.GetPopularVideosByIDs(ctx, videoIDs)
	if err != nil {
		logx.Errorf("batchGetPopulars failed: %v", err)
		populars = map[int64]types.VideoPopular{}
	}

	result := make([]types.VideoPopular, 0, len(videos))
	for _, v := range videos {
		result = append(result, populars[v.VideoID])
	}
	return result
}

// lastPublishedAtMs 解析视频创建时间为毫秒时间戳。
func lastPublishedAtMs(createdAt string) (int64, error) {
	createdAt = strings.TrimSpace(createdAt)
	if createdAt == "" {
		return 0, nil
	}
	t, err := myutils.StrToTime(createdAt, "")
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

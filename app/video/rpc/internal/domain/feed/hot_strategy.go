package feed

import (
	"context"
	"sort"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

// hotFetchLimit Redis 同分过滤兜底：多取若干条再精断。
const hotFetchLimit = 5

// HotStrategy 热门策略：按热度分倒序返回热门视频。
// 热度分主存储在 Redis hot:videos ZSet；Redis 不可用时降级到 MySQL visit_count 兜底。
type HotStrategy struct {
	videoRepo   domainVideoRepo
	popularRepo domainPopularRepo
	feedRepo    domainFeedRepo
}

// NewHotStrategy 创建热门策略。
func NewHotStrategy(videoRepo domainVideoRepo, popularRepo domainPopularRepo, feedRepo domainFeedRepo) *HotStrategy {
	return &HotStrategy{
		videoRepo:   videoRepo,
		popularRepo: popularRepo,
		feedRepo:    feedRepo,
	}
}

// Name 返回策略名。
func (s *HotStrategy) Name() string {
	return "hot"
}

// GetFeed 读取热门 Feed 一页。
func (s *HotStrategy) GetFeed(ctx context.Context, viewerID int64, cursor string, limit int32) (*Result, error) {
	if limit <= 0 || limit > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	parsedCursor, err := DecodeHotCursor(cursor)
	if err != nil {
		return nil, err
	}

	var cursorScore, cursorVideoID int64
	if parsedCursor != nil {
		cursorScore = parsedCursor.Score
		cursorVideoID = parsedCursor.VideoID
	}

	populars, err := s.fetchPopulars(ctx, cursorScore, cursorVideoID, limit)
	if err != nil {
		return nil, err
	}

	hasMore := len(populars) > int(limit)
	if hasMore {
		populars = populars[:limit]
	}

	if len(populars) == 0 {
		return &Result{
			Videos:   nil,
			Populars: nil,
			HasMore:  false,
			Total:    0,
		}, nil
	}

	videoIDs := make([]int64, 0, len(populars))
	for _, p := range populars {
		videoIDs = append(videoIDs, p.VideoID)
	}

	videos, err := s.videoRepo.GetVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	var nextCursor string
	if len(populars) > 0 {
		last := populars[len(populars)-1]
		nextCursor = EncodeHotCursor(&HotCursor{
			Score:   last.HotScore,
			VideoID: last.VideoID,
		})
	}

	return &Result{
		Videos:     videos,
		Populars:   populars,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      int64(len(videos)),
	}, nil
}

// fetchPopulars 优先从 Redis hot:videos 取，失败或为空时走 MySQL visit_count 兜底。
func (s *HotStrategy) fetchPopulars(ctx context.Context, cursorScore, cursorVideoID int64, limit int32) ([]types.VideoPopular, error) {
	if s.feedRepo != nil {
		indexes, err := s.feedRepo.GetHotVideosByCursor(ctx, cursorScore, cursorVideoID, int(limit)+1)
		if err != nil {
			logx.Errorf("get hot videos from redis failed, fallback to db: %v", err)
		} else if len(indexes) > 0 {
			return s.indexesToPopulars(s.filterIndexes(indexes, cursorScore, cursorVideoID, limit+1)), nil
		}
	}

	populars, err := s.popularRepo.GetPopularVideosByCursor(ctx, cursorScore, cursorVideoID, limit+1)
	if err != nil {
		return nil, err
	}
	// 兜底没有独立热度分字段，用 visit_count 作为排序/游标依据。
	for i := range populars {
		if populars[i].HotScore == 0 {
			populars[i].HotScore = populars[i].VisitCount
		}
	}
	return populars, nil
}

// filterIndexes 对 Redis 返回的索引按 (score, video_id) < (cursorScore, cursorVideoID) 精断并截断。
func (s *HotStrategy) filterIndexes(indexes []types.FeedIndex, cursorScore, cursorVideoID int64, limit int32) []types.FeedIndex {
	filtered := make([]types.FeedIndex, 0, len(indexes))
	for _, idx := range indexes {
		if idx.Score < cursorScore || (idx.Score == cursorScore && idx.VideoID < cursorVideoID) || cursorScore == 0 {
			filtered = append(filtered, idx)
			if len(filtered) >= int(limit) {
				break
			}
		}
	}
	return filtered
}

// indexesToPopulars 把 Redis 热度索引转为 VideoPopular，按热度倒序排列。
func (s *HotStrategy) indexesToPopulars(indexes []types.FeedIndex) []types.VideoPopular {
	// Redis 已按 score 倒序返回，但同分精断后可能乱序，需稳定排序
	sort.SliceStable(indexes, func(i, j int) bool {
		if indexes[i].Score != indexes[j].Score {
			return indexes[i].Score > indexes[j].Score
		}
		return indexes[i].VideoID > indexes[j].VideoID
	})

	result := make([]types.VideoPopular, 0, len(indexes))
	for _, idx := range indexes {
		result = append(result, types.VideoPopular{
			VideoID:  idx.VideoID,
			HotScore: idx.Score,
		})
	}
	return result
}

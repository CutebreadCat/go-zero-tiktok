package feed

import (
	"context"
	"sort"

	"go_zero-tiktok/pkg/contract"
	appLogger "go_zero-tiktok/pkg/logger"
	"go_zero-tiktok/pkg/xerr"
)

// hotFetchLimit Redis 同分过滤兜底：多取若干条再精断。
const hotFetchLimit = 5

// hotSelfHealLimit 单次请求最多从榜单移除的失效成员数，避免异常数据一次性清空榜单。
const hotSelfHealLimit = 50

// HotStrategy 严格热度排行榜（scene=hot）。
// 语义边界：唯一的排名数据源是 Redis hot:videos，按热度分倒序一直翻到末尾；
// 取完即结束，不跨榜续接，也不降级到 MySQL visit_count（那是另一套排序，属于"换榜"）。
// 若需要"热度作为排序因子、列表跟时间滚动"的信息流，应走 RecommendStrategy，
// 不要在这里混入兜底或换榜逻辑。
type HotStrategy struct {
	videoRepo domainVideoRepo
	feedRepo  domainHotIndexRepo
}

// NewHotStrategy 创建热门排行榜策略。
func NewHotStrategy(videoRepo domainVideoRepo, feedRepo domainHotIndexRepo) *HotStrategy {
	return &HotStrategy{
		videoRepo: videoRepo,
		feedRepo:  feedRepo,
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

	// 游标按本页实际消费到的最后一条索引推进，即使整页视频都被删除也不会原地打转。
	last := populars[len(populars)-1]

	videos, err := s.videoRepo.GetVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	// GetVideosByIDs 会过滤已删除视频并按传入顺序返回；这里按 video_id 重新对齐热度信息，
	// 避免上层按下标配对时把热度分挂到别的视频上。
	popularByID := make(map[int64]types.VideoPopular, len(populars))
	for _, p := range populars {
		popularByID[p.VideoID] = p
	}
	aligned := make([]types.VideoPopular, 0, len(videos))
	for _, v := range videos {
		if p, ok := popularByID[v.VideoID]; ok {
			aligned = append(aligned, p)
		}
	}
	populars = aligned

	// 详情缺失说明视频已不存在（被删除过滤或命中负缓存），顺手从榜单移除，避免榜单长期出现空洞。
	// GetVideosByIDs 出错会提前返回，走不到这里，所以这里删掉的都是确认查不到的 id。
	s.removeStaleIndexes(ctx, videoIDs, videos)

	nextCursor := EncodeHotCursor(&HotCursor{
		Score:   last.HotScore,
		VideoID: last.VideoID,
	})

	return &Result{
		Videos:     videos,
		Populars:   populars,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      int64(len(videos)),
	}, nil
}

// removeStaleIndexes 把请求了但未能水合出详情的视频从 hot:videos 移除（best-effort，不影响本次响应）。
func (s *HotStrategy) removeStaleIndexes(ctx context.Context, requested []int64, hydrated []types.VideoBaseinfo) {
	if s.feedRepo == nil || len(requested) == 0 {
		return
	}

	returned := make(map[int64]struct{}, len(hydrated))
	for _, v := range hydrated {
		returned[v.VideoID] = struct{}{}
	}

	stale := make([]int64, 0, len(requested))
	for _, id := range requested {
		if _, ok := returned[id]; !ok {
			stale = append(stale, id)
		}
	}
	if len(stale) == 0 {
		return
	}

	total := len(stale)
	if total > hotSelfHealLimit {
		stale = stale[:hotSelfHealLimit]
	}
	appLogger.Warnf("hot leaderboard has %d stale videos, removing %d from index: %v", total, len(stale), stale)
	if err := s.feedRepo.RemoveHotVideos(ctx, stale); err != nil {
		appLogger.Errorf("remove stale hot videos failed: %v", err)
	}
}

// fetchPopulars 只从 hot:videos 读取排行榜。
// 排行榜语义：热度分是唯一排名依据，取完即结束；
// Redis 不可用时返回空列表（错误只进日志），不拿 visit_count 之类的指标冒充热度。
func (s *HotStrategy) fetchPopulars(ctx context.Context, cursorScore, cursorVideoID int64, limit int32) ([]types.VideoPopular, error) {
	if s.feedRepo == nil {
		appLogger.Error("hot leaderboard index is not configured")
		return nil, nil
	}

	indexes, err := s.feedRepo.GetHotVideosByCursor(ctx, cursorScore, cursorVideoID, int(limit)+1)
	if err != nil {
		// 排行榜不可用时返回空列表，由前端展示"暂无内容"；错误只留在日志里，不冒充其它榜单。
		appLogger.Errorf("get hot leaderboard failed, cursorScore=%d, cursorVideoID=%d: %v",
			cursorScore, cursorVideoID, err)
		return nil, nil
	}

	if len(indexes) == 0 {
		// 首页为空说明索引异常（写入失败或 key 被清空），翻页为空则是正常翻到底，两者都要可观测。
		if cursorScore == 0 {
			appLogger.Error("hot leaderboard is empty on first page, check hot:videos index")
		} else {
			appLogger.Infof("hot leaderboard reached the end, cursorScore=%d", cursorScore)
		}
		return nil, nil
	}

	return s.indexesToPopulars(s.filterIndexes(indexes, cursorScore, cursorVideoID, limit+1)), nil
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

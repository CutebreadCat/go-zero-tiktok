package feed

import (
	"context"
	"sort"
	"time"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

// RecommendConfig 推荐策略配置。
type RecommendConfig struct {
	// 召回倍数：每路召回 limit * FetchFactor 条候选
	FetchFactor int
	// 单作者每页最大重复次数
	MaxAuthorRepeat int
	// 曝光记录 TTL
	SeenTTL time.Duration
	// 曝光记录最大容量
	SeenMaxSize int
	// 打分权重
	Weights ScorerWeights
}

// DefaultRecommendConfig 返回默认推荐配置。
func DefaultRecommendConfig() RecommendConfig {
	return RecommendConfig{
		FetchFactor:     3,
		MaxAuthorRepeat: 2,
		SeenTTL:         7 * 24 * time.Hour,
		SeenMaxSize:     5000,
		Weights:         DefaultScorerWeights(),
	}
}

// recommendCandidate 是推荐策略内部候选。
type recommendCandidate struct {
	video    types.VideoBaseinfo
	popular  types.VideoPopular
	score    int64
	followed bool
	source   string // inbox / hot / global / fallback
}

// RecommendStrategy 推荐策略：融合关注/热门/时间线三路召回，规则打分，去重打散。
type RecommendStrategy struct {
	videoRepo   domainVideoRepo
	popularRepo domainPopularRepo
	feedRepo    domainFeedRepo
	qosRepo     domainQoSRepo
	seenRepo    domainSeenRepo
	scorer      *Scorer
	config      RecommendConfig
}

// NewRecommendStrategy 创建推荐策略。
func NewRecommendStrategy(
	videoRepo domainVideoRepo,
	popularRepo domainPopularRepo,
	feedRepo domainFeedRepo,
	qosRepo domainQoSRepo,
	seenRepo domainSeenRepo,
	config RecommendConfig,
) *RecommendStrategy {
	if config.FetchFactor <= 0 {
		config.FetchFactor = DefaultRecommendConfig().FetchFactor
	}
	if config.MaxAuthorRepeat <= 0 {
		config.MaxAuthorRepeat = DefaultRecommendConfig().MaxAuthorRepeat
	}
	if config.SeenTTL <= 0 {
		config.SeenTTL = DefaultRecommendConfig().SeenTTL
	}
	if config.SeenMaxSize <= 0 {
		config.SeenMaxSize = DefaultRecommendConfig().SeenMaxSize
	}

	return &RecommendStrategy{
		videoRepo:   videoRepo,
		popularRepo: popularRepo,
		feedRepo:    feedRepo,
		qosRepo:     qosRepo,
		seenRepo:    seenRepo,
		scorer:      NewScorer(config.Weights),
		config:      config,
	}
}

// Name 返回策略名。
func (s *RecommendStrategy) Name() string {
	return "recommend"
}

// GetFeed 读取推荐 Feed 一页。
func (s *RecommendStrategy) GetFeed(ctx context.Context, viewerID int64, cursor string, limit int32) (*Result, error) {
	if limit <= 0 || limit > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	parsedCursor, err := DecodeRecommendCursor(cursor)
	if err != nil {
		return nil, err
	}

	// 1. 多源召回
	candidates, err := s.recall(ctx, viewerID, limit)
	if err != nil {
		return nil, err
	}

	// 2. 粗排打分
	scored := s.score(ctx, candidates, viewerID)

	// 3. 按推荐分倒序（同分按 video_id 降序，保证顺序稳定）
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].video.VideoID > scored[j].video.VideoID
	})

	// 4. 去重 + 打散 + 截断
	selected := s.rerank(ctx, viewerID, scored, parsedCursor, limit)

	// 5. 补量兜底
	if len(selected) < int(limit) {
		selected = s.fallback(ctx, viewerID, selected, parsedCursor, limit)
	}

	// 6. 构造返回
	videos := make([]types.VideoBaseinfo, 0, len(selected))
	populars := make([]types.VideoPopular, 0, len(selected))
	seenVideoIDs := make([]int64, 0, len(selected))
	for _, c := range selected {
		videos = append(videos, c.video)
		populars = append(populars, c.popular)
		seenVideoIDs = append(seenVideoIDs, c.video.VideoID)
	}

	// 7. 写入曝光记录（失败不阻断主链路）
	if s.seenRepo != nil && viewerID > 0 && len(seenVideoIDs) > 0 {
		if err := s.seenRepo.MarkSeen(ctx, viewerID, seenVideoIDs); err != nil {
			logx.Errorf("mark seen failed for user %d: %v", viewerID, err)
		}
	}

	// 8. 构造游标
	var nextCursor string
	if len(selected) > 0 {
		last := selected[len(selected)-1]
		nextCursor = EncodeRecommendCursor(&RecommendCursor{
			Score:   last.score,
			VideoID: last.video.VideoID,
		})
	}

	// hasMore：取满 limit 且打分队列里还有未被选中的候选，则认为有下一页
	hasMore := len(selected) == int(limit) && len(scored) > len(selected)

	return &Result{
		Videos:     videos,
		Populars:   populars,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      int64(len(videos)),
	}, nil
}

// recall 多源召回候选。
func (s *RecommendStrategy) recall(ctx context.Context, viewerID int64, limit int32) ([]recommendCandidate, error) {
	fetchCount := int(limit) * s.config.FetchFactor
	if fetchCount < int(limit)*2 {
		fetchCount = int(limit) * 2
	}

	sources := []struct {
		name    string
		indexes []types.FeedIndex
	}{
		{name: "inbox"},
		{name: "hot"},
		{name: "global"},
	}

	// 关注流召回
	if viewerID > 0 && s.feedRepo != nil {
		inbox, err := s.feedRepo.GetUserInbox(ctx, viewerID, 0, fetchCount)
		if err != nil {
			logx.Errorf("recommend recall inbox failed: %v", err)
		} else {
			sources[0].indexes = inbox
		}
	}

	// 热门池召回
	if s.feedRepo != nil {
		hot, err := s.feedRepo.GetHotVideosByCursor(ctx, 0, 0, fetchCount)
		if err != nil {
			logx.Errorf("recommend recall hot failed: %v", err)
		} else {
			sources[1].indexes = hot
		}
	}

	// 全站时间线召回
	if s.feedRepo != nil {
		global, err := s.feedRepo.GetGlobalPool(ctx, 0, fetchCount)
		if err != nil {
			logx.Errorf("recommend recall global failed: %v", err)
		} else {
			sources[2].indexes = global
		}
	}

	// 合并索引并记录来源
	sourceMap := make(map[int64]string)
	allIndexes := make([]types.FeedIndex, 0, fetchCount*3)
	for _, src := range sources {
		for _, idx := range src.indexes {
			// 优先保留第一次出现的来源（inbox 优先级最高）
			if _, ok := sourceMap[idx.VideoID]; !ok {
				sourceMap[idx.VideoID] = src.name
			}
			allIndexes = append(allIndexes, idx)
		}
	}

	// 去重 video_id
	seen := make(map[int64]struct{}, len(allIndexes))
	uniqueIDs := make([]int64, 0, len(allIndexes))
	for _, idx := range allIndexes {
		if _, ok := seen[idx.VideoID]; ok {
			continue
		}
		seen[idx.VideoID] = struct{}{}
		uniqueIDs = append(uniqueIDs, idx.VideoID)
	}

	if len(uniqueIDs) == 0 {
		return nil, nil
	}

	// 水合视频信息
	videos, err := s.videoRepo.GetVideosByIDs(ctx, uniqueIDs)
	if err != nil {
		return nil, err
	}

	candidates := make([]recommendCandidate, 0, len(videos))
	for _, v := range videos {
		candidates = append(candidates, recommendCandidate{
			video:  v,
			source: sourceMap[v.VideoID],
		})
	}
	return candidates, nil
}

// score 批量打分。
func (s *RecommendStrategy) score(ctx context.Context, candidates []recommendCandidate, viewerID int64) []recommendCandidate {
	if len(candidates) == 0 {
		return nil
	}

	videoIDs := make([]int64, 0, len(candidates))
	for _, c := range candidates {
		videoIDs = append(videoIDs, c.video.VideoID)
	}

	// 批量查热度
	populars, err := s.popularRepo.GetPopularVideosByIDs(ctx, videoIDs)
	if err != nil {
		logx.Errorf("recommend score get populars failed: %v", err)
		populars = map[int64]types.VideoPopular{}
	}

	// 批量查 QoS
	var qosMetrics map[int64]types.VideoQoSMetrics
	if s.qosRepo != nil {
		qosMetrics, err = s.qosRepo.GetQoSMetricsByVideoIDs(ctx, videoIDs)
		if err != nil {
			logx.Errorf("recommend score get qos failed: %v", err)
			qosMetrics = map[int64]types.VideoQoSMetrics{}
		}
	}

	now := time.Now()
	result := make([]recommendCandidate, 0, len(candidates))
	for i := range candidates {
		c := &candidates[i]
		c.popular = populars[c.video.VideoID]
		c.popular.VideoQoSMetrics = qosMetrics[c.video.VideoID]
		// inbox 来源的视频作者视为已关注
		c.followed = c.source == "inbox"
		c.score = s.scorer.Score(c.video, c.popular, c.followed, now)
		result = append(result, *c)
	}
	return result
}

// rerank 去重 + 打散 + 截断。
func (s *RecommendStrategy) rerank(ctx context.Context, viewerID int64, scored []recommendCandidate, cursor *RecommendCursor, limit int32) []recommendCandidate {
	selected := make([]recommendCandidate, 0, limit)
	authorCount := make(map[int64]int)

	for _, c := range scored {
		// 游标过滤：只取比游标分数低的，同分按 video_id 精断
		if cursor != nil {
			if c.score > cursor.Score {
				continue
			}
			if c.score == cursor.Score && c.video.VideoID >= cursor.VideoID {
				continue
			}
		}

		// 曝光去重
		if s.isSeen(ctx, viewerID, c.video.VideoID) {
			continue
		}

		// 同作者打散
		if authorCount[c.video.AuthorID] >= s.config.MaxAuthorRepeat {
			continue
		}

		selected = append(selected, c)
		authorCount[c.video.AuthorID]++

		if len(selected) >= int(limit) {
			break
		}
	}

	return selected
}

// isSeen 判断视频是否已曝光。
func (s *RecommendStrategy) isSeen(ctx context.Context, viewerID, videoID int64) bool {
	if s.seenRepo == nil || viewerID <= 0 {
		return false
	}
	seen, err := s.seenRepo.IsSeen(ctx, viewerID, videoID)
	if err != nil {
		logx.Errorf("check seen failed for user %d video %d: %v", viewerID, videoID, err)
		return false
	}
	return seen
}

// fallback 从时间线兜底补量。
func (s *RecommendStrategy) fallback(ctx context.Context, viewerID int64, selected []recommendCandidate, cursor *RecommendCursor, limit int32) []recommendCandidate {
	need := int(limit) - len(selected)
	if need <= 0 {
		return selected
	}

	// 兜底取最新发布的时间线视频
	videos, err := s.videoRepo.GetVideosByCursor(ctx, 0, 0, int32(need*2))
	if err != nil {
		logx.Errorf("recommend fallback failed: %v", err)
		return selected
	}

	populars := batchGetPopulars(ctx, videos, s.popularRepo)
	now := time.Now()
	existing := make(map[int64]struct{}, len(selected))
	authorCount := make(map[int64]int)
	for _, c := range selected {
		existing[c.video.VideoID] = struct{}{}
		authorCount[c.video.AuthorID]++
	}

	for i, v := range videos {
		if _, ok := existing[v.VideoID]; ok {
			continue
		}
		if s.isSeen(ctx, viewerID, v.VideoID) {
			continue
		}
		if authorCount[v.AuthorID] >= s.config.MaxAuthorRepeat {
			continue
		}
		var popular types.VideoPopular
		if i < len(populars) {
			popular = populars[i]
		}
		popular.VideoID = v.VideoID
		selected = append(selected, recommendCandidate{
			video:   v,
			popular: popular,
			score:   s.scorer.Score(v, popular, false, now),
			source:  "fallback",
		})
		authorCount[v.AuthorID]++
		if len(selected) >= int(limit) {
			break
		}
	}

	return selected
}

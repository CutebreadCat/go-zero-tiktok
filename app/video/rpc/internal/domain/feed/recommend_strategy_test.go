package feed

import (
	"context"
	"testing"

	"go_zero-tiktok/pkg/contract"
)

// mockRecommendVideoRepo 视频仓库 mock。
type mockRecommendVideoRepo struct {
	videos map[int64]types.VideoBaseinfo
}

func (m *mockRecommendVideoRepo) GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error) {
	result := make([]types.VideoBaseinfo, 0, len(videoIDs))
	for _, id := range videoIDs {
		if v, ok := m.videos[id]; ok {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockRecommendVideoRepo) GetVideosByCursor(ctx context.Context, publishedAt, videoID int64, limit int32) ([]types.VideoBaseinfo, error) {
	return nil, nil
}

// mockRecommendPopularRepo 热度仓库 mock。
type mockRecommendPopularRepo struct {
	populars map[int64]types.VideoPopular
}

func (m *mockRecommendPopularRepo) GetPopularVideosByIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoPopular, error) {
	return m.populars, nil
}

func (m *mockRecommendPopularRepo) GetPopularVideosByCursor(ctx context.Context, score, videoID int64, limit int32) ([]types.VideoPopular, error) {
	return nil, nil
}

// mockRecommendFeedRepo Feed 索引仓库 mock。
type mockRecommendFeedRepo struct {
	global []types.FeedIndex
	hot    []types.FeedIndex
	inbox  []types.FeedIndex
}

func (m *mockRecommendFeedRepo) GetGlobalPool(ctx context.Context, lastTimeMs int64, limit int) ([]types.FeedIndex, error) {
	return m.global, nil
}

func (m *mockRecommendFeedRepo) GetUserInbox(ctx context.Context, uid, lastTimeMs int64, limit int) ([]types.FeedIndex, error) {
	return m.inbox, nil
}

func (m *mockRecommendFeedRepo) GetHotVideosByCursor(ctx context.Context, cursorScore, cursorVideoID int64, limit int) ([]types.FeedIndex, error) {
	return m.hot, nil
}

// mockRecommendQoSRepo QoS 仓库 mock。
type mockRecommendQoSRepo struct{}

func (m *mockRecommendQoSRepo) GetQoSMetricsByVideoIDs(ctx context.Context, videoIDs []int64) (map[int64]types.VideoQoSMetrics, error) {
	return map[int64]types.VideoQoSMetrics{}, nil
}

// mockRecommendSeenRepo 曝光仓库 mock。
type mockRecommendSeenRepo struct {
	seen map[int64]map[int64]bool
}

func (m *mockRecommendSeenRepo) IsSeen(ctx context.Context, userID, videoID int64) (bool, error) {
	if m.seen[userID] == nil {
		return false, nil
	}
	return m.seen[userID][videoID], nil
}

func (m *mockRecommendSeenRepo) MarkSeen(ctx context.Context, userID int64, videoIDs []int64) error {
	if m.seen[userID] == nil {
		m.seen[userID] = make(map[int64]bool)
	}
	for _, id := range videoIDs {
		m.seen[userID][id] = true
	}
	return nil
}

func TestRecommendStrategy_GetFeed_Empty(t *testing.T) {
	strategy := NewRecommendStrategy(
		&mockRecommendVideoRepo{},
		&mockRecommendPopularRepo{},
		&mockRecommendFeedRepo{},
		&mockRecommendQoSRepo{},
		&mockRecommendSeenRepo{seen: make(map[int64]map[int64]bool)},
		DefaultRecommendConfig(),
	)

	result, err := strategy.GetFeed(context.Background(), 1, "", 10)
	if err != nil {
		t.Fatalf("GetFeed() error = %v", err)
	}
	if len(result.Videos) != 0 {
		t.Errorf("GetFeed() videos = %d, want 0", len(result.Videos))
	}
	if result.HasMore {
		t.Error("GetFeed() HasMore = true, want false")
	}
}

func TestRecommendStrategy_GetFeed_Dedup(t *testing.T) {
	nowStr := "2026-08-23 12:00:00"
	videos := map[int64]types.VideoBaseinfo{
		1: {VideoID: 1, AuthorID: 100, CreatedAt: nowStr},
		2: {VideoID: 2, AuthorID: 101, CreatedAt: nowStr},
		3: {VideoID: 3, AuthorID: 102, CreatedAt: nowStr},
	}
	populars := map[int64]types.VideoPopular{
		1: {VideoID: 1, HotScore: 100},
		2: {VideoID: 2, HotScore: 200},
		3: {VideoID: 3, HotScore: 300},
	}

	seen := &mockRecommendSeenRepo{
		seen: map[int64]map[int64]bool{
			1: {2: true}, // 用户 1 已看过视频 2
		},
	}

	strategy := NewRecommendStrategy(
		&mockRecommendVideoRepo{videos: videos},
		&mockRecommendPopularRepo{populars: populars},
		&mockRecommendFeedRepo{
			global: []types.FeedIndex{{VideoID: 1}, {VideoID: 2}, {VideoID: 3}},
		},
		&mockRecommendQoSRepo{},
		seen,
		DefaultRecommendConfig(),
	)

	result, err := strategy.GetFeed(context.Background(), 1, "", 10)
	if err != nil {
		t.Fatalf("GetFeed() error = %v", err)
	}

	// 视频 2 已被曝光，不应出现在结果中
	for _, v := range result.Videos {
		if v.VideoID == 2 {
			t.Error("GetFeed() returned already seen video 2")
		}
	}

	// 写入曝光后，视频 1 和 3 应被标记
	if !seen.seen[1][1] || !seen.seen[1][3] {
		t.Error("GetFeed() did not mark returned videos as seen")
	}
}

func TestRecommendStrategy_GetFeed_AuthorDiversity(t *testing.T) {
	nowStr := "2026-08-23 12:00:00"
	videos := map[int64]types.VideoBaseinfo{
		1: {VideoID: 1, AuthorID: 100, CreatedAt: nowStr},
		2: {VideoID: 2, AuthorID: 100, CreatedAt: nowStr},
		3: {VideoID: 3, AuthorID: 100, CreatedAt: nowStr},
		4: {VideoID: 4, AuthorID: 101, CreatedAt: nowStr},
	}
	populars := map[int64]types.VideoPopular{
		1: {VideoID: 1, HotScore: 400},
		2: {VideoID: 2, HotScore: 300},
		3: {VideoID: 3, HotScore: 200},
		4: {VideoID: 4, HotScore: 100},
	}

	strategy := NewRecommendStrategy(
		&mockRecommendVideoRepo{videos: videos},
		&mockRecommendPopularRepo{populars: populars},
		&mockRecommendFeedRepo{
			global: []types.FeedIndex{{VideoID: 1}, {VideoID: 2}, {VideoID: 3}, {VideoID: 4}},
		},
		&mockRecommendQoSRepo{},
		&mockRecommendSeenRepo{seen: make(map[int64]map[int64]bool)},
		RecommendConfig{
			FetchFactor:     2,
			MaxAuthorRepeat: 2,
			SeenTTL:         DefaultRecommendConfig().SeenTTL,
			SeenMaxSize:     DefaultRecommendConfig().SeenMaxSize,
			Weights:         DefaultScorerWeights(),
		},
	)

	result, err := strategy.GetFeed(context.Background(), 1, "", 2)
	if err != nil {
		t.Fatalf("GetFeed() error = %v", err)
	}
	if len(result.Videos) != 2 {
		t.Fatalf("GetFeed() videos = %d, want 2", len(result.Videos))
	}

	// 每页最多 2 条同作者，所以前两条应该是视频 1 和 2（同作者 100）
	// 但视频 4（作者 101）应该排在第三位，这里 limit=2，所以结果里最多只有作者 100 的视频
	authorCount := make(map[int64]int)
	for _, v := range result.Videos {
		authorCount[v.AuthorID]++
	}
	for authorID, count := range authorCount {
		if count > 2 {
			t.Errorf("author %d appeared %d times, max 2", authorID, count)
		}
	}
}

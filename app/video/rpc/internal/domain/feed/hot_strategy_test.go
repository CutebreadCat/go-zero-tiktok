package feed

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/pkg/contract"
)

// mockHotIndexRepo 热门索引仓储 mock。
type mockHotIndexRepo struct {
	indexes []types.FeedIndex
	err     error

	gotCursorScore   int64
	gotCursorVideoID int64
	gotLimit         int

	removed [][]int64
}

func (m *mockHotIndexRepo) GetHotVideosByCursor(ctx context.Context, cursorScore, cursorVideoID int64, limit int) ([]types.FeedIndex, error) {
	m.gotCursorScore = cursorScore
	m.gotCursorVideoID = cursorVideoID
	m.gotLimit = limit
	if m.err != nil {
		return nil, m.err
	}
	return m.indexes, nil
}

func (m *mockHotIndexRepo) RemoveHotVideos(ctx context.Context, videoIDs []int64) error {
	ids := append([]int64(nil), videoIDs...)
	m.removed = append(m.removed, ids)
	return nil
}

// mockHotVideoRepo 视频详情仓储 mock：按传入顺序返回，缺失的 id 自动跳过，模拟 MySQL 过滤已删除视频。
type mockHotVideoRepo struct {
	videos  map[int64]types.VideoBaseinfo
	calls   int
	lastIDs []int64
}

func (m *mockHotVideoRepo) GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error) {
	m.calls++
	m.lastIDs = append([]int64(nil), videoIDs...)
	result := make([]types.VideoBaseinfo, 0, len(videoIDs))
	for _, id := range videoIDs {
		if v, ok := m.videos[id]; ok {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockHotVideoRepo) GetVideosByCursor(ctx context.Context, publishedAt, videoID int64, limit int32) ([]types.VideoBaseinfo, error) {
	return nil, nil
}

func TestHotStrategy_GetFeed_FirstPage(t *testing.T) {
	feedRepo := &mockHotIndexRepo{indexes: []types.FeedIndex{
		{VideoID: 1, Score: 100000},
		{VideoID: 2, Score: 98000},
		{VideoID: 3, Score: 96000},
	}}
	videoRepo := &mockHotVideoRepo{videos: map[int64]types.VideoBaseinfo{
		1: {VideoID: 1},
		2: {VideoID: 2},
		3: {VideoID: 3},
	}}
	s := NewHotStrategy(videoRepo, feedRepo)

	res, err := s.GetFeed(context.Background(), 0, "", 2)
	if err != nil {
		t.Fatalf("GetFeed() error = %v", err)
	}

	if len(res.Videos) != 2 || len(res.Populars) != 2 {
		t.Fatalf("got videos=%d populars=%d, want 2/2", len(res.Videos), len(res.Populars))
	}
	for i := range res.Videos {
		if res.Videos[i].VideoID != res.Populars[i].VideoID {
			t.Errorf("index %d: video_id=%d popular_video_id=%d", i, res.Videos[i].VideoID, res.Populars[i].VideoID)
		}
	}
	if !res.HasMore {
		t.Error("HasMore = false, want true")
	}

	cursor, err := DecodeHotCursor(res.NextCursor)
	if err != nil {
		t.Fatalf("DecodeHotCursor() error = %v", err)
	}
	if cursor == nil || cursor.Score != 98000 || cursor.VideoID != 2 {
		t.Errorf("next cursor = %+v, want score 98000 / video 2", cursor)
	}
	if feedRepo.gotCursorScore != 0 || feedRepo.gotCursorVideoID != 0 {
		t.Errorf("repo cursor = (%d,%d), want first page (0,0)", feedRepo.gotCursorScore, feedRepo.gotCursorVideoID)
	}
}

func TestHotStrategy_GetFeed_EndOfLeaderboard(t *testing.T) {
	feedRepo := &mockHotIndexRepo{indexes: nil}
	videoRepo := &mockHotVideoRepo{videos: map[int64]types.VideoBaseinfo{}}
	s := NewHotStrategy(videoRepo, feedRepo)

	res, err := s.GetFeed(context.Background(), 0, EncodeHotCursor(&HotCursor{Score: 100, VideoID: 9}), 5)
	if err != nil {
		t.Fatalf("GetFeed() error = %v", err)
	}
	if len(res.Videos) != 0 || len(res.Populars) != 0 {
		t.Errorf("got videos=%d populars=%d, want empty", len(res.Videos), len(res.Populars))
	}
	if res.HasMore {
		t.Error("HasMore = true, want false at end of leaderboard")
	}
	if videoRepo.calls != 0 {
		t.Errorf("videoRepo called %d times, want 0 (no fallback to another ranking)", videoRepo.calls)
	}
}

func TestHotStrategy_GetFeed_IndexUnavailable(t *testing.T) {
	feedRepo := &mockHotIndexRepo{err: errors.New("redis down")}
	videoRepo := &mockHotVideoRepo{videos: map[int64]types.VideoBaseinfo{}}
	s := NewHotStrategy(videoRepo, feedRepo)

	res, err := s.GetFeed(context.Background(), 0, "", 5)
	if err != nil {
		t.Fatalf("GetFeed() error = %v, want nil (empty list)", err)
	}
	if len(res.Videos) != 0 {
		t.Errorf("got %d videos, want empty", len(res.Videos))
	}
	if videoRepo.calls != 0 {
		t.Errorf("videoRepo called %d times, want 0", videoRepo.calls)
	}
}

func TestHotStrategy_GetFeed_AlignsDeletedVideo(t *testing.T) {
	feedRepo := &mockHotIndexRepo{indexes: []types.FeedIndex{
		{VideoID: 1, Score: 100000},
		{VideoID: 2, Score: 98000},
		{VideoID: 3, Score: 96000},
	}}
	// 视频 2 已删除：GetVideosByIDs 只返回 1 和 3。
	videoRepo := &mockHotVideoRepo{videos: map[int64]types.VideoBaseinfo{
		1: {VideoID: 1},
		3: {VideoID: 3},
	}}
	s := NewHotStrategy(videoRepo, feedRepo)

	res, err := s.GetFeed(context.Background(), 0, "", 3)
	if err != nil {
		t.Fatalf("GetFeed() error = %v", err)
	}

	if len(res.Videos) != 2 || len(res.Populars) != 2 {
		t.Fatalf("got videos=%d populars=%d, want 2/2", len(res.Videos), len(res.Populars))
	}
	wantScores := map[int64]int64{1: 100000, 3: 96000}
	for i := range res.Videos {
		vid := res.Videos[i].VideoID
		if res.Populars[i].VideoID != vid {
			t.Errorf("index %d: video_id=%d popular_video_id=%d", i, vid, res.Populars[i].VideoID)
		}
		if res.Populars[i].HotScore != wantScores[vid] {
			t.Errorf("video %d hot_score=%d, want %d", vid, res.Populars[i].HotScore, wantScores[vid])
		}
	}

	if len(feedRepo.removed) != 1 || len(feedRepo.removed[0]) != 1 || feedRepo.removed[0][0] != 2 {
		t.Errorf("removed = %v, want [[2]]", feedRepo.removed)
	}
}

func TestHotStrategy_GetFeed_UsesCursor(t *testing.T) {
	feedRepo := &mockHotIndexRepo{indexes: nil}
	videoRepo := &mockHotVideoRepo{videos: map[int64]types.VideoBaseinfo{}}
	s := NewHotStrategy(videoRepo, feedRepo)

	cursor := EncodeHotCursor(&HotCursor{Score: 98000, VideoID: 2})
	if _, err := s.GetFeed(context.Background(), 0, cursor, 5); err != nil {
		t.Fatalf("GetFeed() error = %v", err)
	}

	if feedRepo.gotCursorScore != 98000 || feedRepo.gotCursorVideoID != 2 {
		t.Errorf("repo cursor = (%d,%d), want (98000,2)", feedRepo.gotCursorScore, feedRepo.gotCursorVideoID)
	}
}

func TestHotStrategy_removeStaleIndexes_Capped(t *testing.T) {
	feedRepo := &mockHotIndexRepo{}
	s := NewHotStrategy(&mockHotVideoRepo{}, feedRepo)

	requested := make([]int64, 60)
	for i := range requested {
		requested[i] = int64(i + 1)
	}

	s.removeStaleIndexes(context.Background(), requested, nil)

	if len(feedRepo.removed) != 1 {
		t.Fatalf("RemoveHotVideos called %d times, want 1", len(feedRepo.removed))
	}
	if len(feedRepo.removed[0]) != hotSelfHealLimit {
		t.Errorf("removed %d ids, want capped at %d", len(feedRepo.removed[0]), hotSelfHealLimit)
	}
}

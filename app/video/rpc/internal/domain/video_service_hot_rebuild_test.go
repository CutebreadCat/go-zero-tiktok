package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	feedpkg "go_zero-tiktok/app/video/rpc/internal/domain/feed"
	"go_zero-tiktok/pkg/contract"
)

// fakePopularRepo 按 keyset 语义从预置快照中分页，模拟 MySQL 游标扫描。
type fakePopularRepo struct {
	IPopularRepo

	rows     []types.VideoStatSnapshot
	count    int64
	listErr  error
	countErr error

	listCalls int
	lastSince time.Time
}

func (f *fakePopularRepo) ListVideoStatsForRebuild(_ context.Context, since, cursorAt time.Time, cursorID int64, limit int32) ([]types.VideoStatSnapshot, error) {
	f.listCalls++
	f.lastSince = since
	if f.listErr != nil {
		return nil, f.listErr
	}

	result := make([]types.VideoStatSnapshot, 0, limit)
	for _, row := range f.rows {
		if !since.IsZero() && row.PublishAt.Before(since) {
			continue
		}
		if !cursorAt.IsZero() {
			if row.PublishAt.Before(cursorAt) {
				continue
			}
			if row.PublishAt.Equal(cursorAt) && row.VideoID <= cursorID {
				continue
			}
		} else if row.VideoID <= cursorID {
			continue
		}

		result = append(result, row)
		if int32(len(result)) >= limit {
			break
		}
	}
	return result, nil
}

func (f *fakePopularRepo) CountVideoStatsForRebuild(_ context.Context, since time.Time) (int64, error) {
	f.lastSince = since
	return f.count, f.countErr
}

// fakeFeedRepo 记录重建期间的锁、临时键与发布动作。
type fakeFeedRepo struct {
	IFeedRepo

	lockFail   bool
	appendErr  error
	publishErr error
	members    int64

	locked    bool
	unlocked  bool
	resets    int
	appended  []types.FeedIndex
	published bool
	keepTopN  int
}

func (f *fakeFeedRepo) TryLockHotRebuild(_ context.Context, _ time.Duration) (string, bool, error) {
	if f.lockFail {
		return "", false, nil
	}
	f.locked = true
	return "token-1", true, nil
}

func (f *fakeFeedRepo) UnlockHotRebuild(_ context.Context, _ string) error {
	f.unlocked = true
	return nil
}

func (f *fakeFeedRepo) ResetHotRebuild(_ context.Context) error {
	f.resets++
	return nil
}

func (f *fakeFeedRepo) AppendHotRebuildScores(_ context.Context, scores []types.FeedIndex) error {
	if f.appendErr != nil {
		return f.appendErr
	}
	f.appended = append(f.appended, scores...)
	return nil
}

func (f *fakeFeedRepo) PublishHotRebuild(_ context.Context, keepTopN int) (int64, error) {
	if f.publishErr != nil {
		return 0, f.publishErr
	}
	f.published = true
	f.keepTopN = keepTopN
	return f.members, nil
}

func newRebuildTestService(popular *fakePopularRepo, feedRepo *fakeFeedRepo) *VideoService {
	calc := &feedpkg.HotScoreCalculator{
		BaseScore:      100,
		LikeWeight:     5,
		CommentWeight:  10,
		FavoriteWeight: 3,
		Gravity:        1.5,
	}
	return NewVideoService(nil, popular, nil, nil, feedRepo, nil, calc, nil, feedpkg.RecommendConfig{})
}

func rebuildRows(n int, publishAt time.Time) []types.VideoStatSnapshot {
	rows := make([]types.VideoStatSnapshot, 0, n)
	for i := 1; i <= n; i++ {
		rows = append(rows, types.VideoStatSnapshot{
			VideoID:       int64(i),
			VisitCount:    int64(i * 10),
			LikeCount:     int64(i),
			CommentCount:  int64(i),
			FavoriteCount: int64(i),
			PublishAt:     publishAt.Add(time.Duration(i) * time.Minute),
		})
	}
	return rows
}

// TestRebuildHotRanking_BatchesAndPublishes 分批扫描全部候选，写入同一临时键后一次性发布。
func TestRebuildHotRanking_BatchesAndPublishes(t *testing.T) {
	publishAt := time.Now().Add(-24 * time.Hour)
	popular := &fakePopularRepo{rows: rebuildRows(5, publishAt)}
	feedRepo := &fakeFeedRepo{members: 3}
	svc := newRebuildTestService(popular, feedRepo)

	scanned, members, err := svc.RebuildHotRanking(context.Background(), HotRebuildOptions{BatchSize: 2, KeepTopN: 3})
	if err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}

	if scanned != 5 {
		t.Fatalf("scanned = %d, want 5", scanned)
	}
	if members != 3 {
		t.Fatalf("members = %d, want 3", members)
	}
	if len(feedRepo.appended) != 5 {
		t.Fatalf("appended = %d, want 5", len(feedRepo.appended))
	}
	// 2 + 2 + 1 三批数据，再加一次空页确认结束
	if popular.listCalls != 4 {
		t.Fatalf("list calls = %d, want 4", popular.listCalls)
	}
	if !feedRepo.published {
		t.Fatal("expected hot ranking published")
	}
	if feedRepo.keepTopN != 3 {
		t.Fatalf("keep_top_n = %d, want 3", feedRepo.keepTopN)
	}
	if !feedRepo.unlocked {
		t.Fatal("expected rebuild lock released")
	}
	if feedRepo.resets != 1 {
		t.Fatalf("reset calls = %d, want 1 (initial clean)", feedRepo.resets)
	}

	// 分值必须与实时重算公式一致
	want := svc.hotScoreCalculator.Compute(types.VideoPopular{
		VideoID:       3,
		VisitCount:    30,
		LikeCount:     3,
		CommentCount:  3,
		FavoriteCount: 3,
	}, popular.rows[2].PublishAt)
	if feedRepo.appended[2].Score != want {
		t.Fatalf("score = %d, want %d", feedRepo.appended[2].Score, want)
	}
}

// TestRebuildHotRanking_DryRun 预检只统计候选数，不抢锁也不触碰 Redis。
func TestRebuildHotRanking_DryRun(t *testing.T) {
	popular := &fakePopularRepo{count: 42}
	feedRepo := &fakeFeedRepo{}
	svc := newRebuildTestService(popular, feedRepo)

	scanned, members, err := svc.RebuildHotRanking(context.Background(), HotRebuildOptions{DryRun: true})
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if scanned != 42 || members != 0 {
		t.Fatalf("scanned/members = %d/%d, want 42/0", scanned, members)
	}
	if feedRepo.locked || feedRepo.resets != 0 || feedRepo.published {
		t.Fatal("dry-run must not lock or write redis")
	}
}

// TestRebuildHotRanking_EmptyResultAborts 空候选不覆盖线上榜，并释放锁。
func TestRebuildHotRanking_EmptyResultAborts(t *testing.T) {
	popular := &fakePopularRepo{}
	feedRepo := &fakeFeedRepo{}
	svc := newRebuildTestService(popular, feedRepo)

	_, _, err := svc.RebuildHotRanking(context.Background(), HotRebuildOptions{})
	if err == nil {
		t.Fatal("expected error on empty candidates")
	}
	if feedRepo.published {
		t.Fatal("empty rebuild must not publish")
	}
	if !feedRepo.unlocked {
		t.Fatal("expected rebuild lock released")
	}
	// 1 次开工前清理 + 1 次失败后清理
	if feedRepo.resets != 2 {
		t.Fatalf("reset calls = %d, want 2 (initial + cleanup)", feedRepo.resets)
	}
}

// TestRebuildHotRanking_LockContention 已有重建在跑时快速失败，不清理别人的临时键。
func TestRebuildHotRanking_LockContention(t *testing.T) {
	popular := &fakePopularRepo{rows: rebuildRows(2, time.Now())}
	feedRepo := &fakeFeedRepo{lockFail: true}
	svc := newRebuildTestService(popular, feedRepo)

	_, _, err := svc.RebuildHotRanking(context.Background(), HotRebuildOptions{})
	if err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("expected already-running error, got %v", err)
	}
	if popular.listCalls != 0 || feedRepo.resets != 0 || feedRepo.published {
		t.Fatal("lock miss must not scan, reset or publish")
	}
}

// TestRebuildHotRanking_AppendFailureCleansUp 中途写临时键失败时清理并解锁，不发布。
func TestRebuildHotRanking_AppendFailureCleansUp(t *testing.T) {
	popular := &fakePopularRepo{rows: rebuildRows(2, time.Now())}
	feedRepo := &fakeFeedRepo{appendErr: errors.New("redis down")}
	svc := newRebuildTestService(popular, feedRepo)

	_, _, err := svc.RebuildHotRanking(context.Background(), HotRebuildOptions{})
	if err == nil {
		t.Fatal("expected append error")
	}
	if feedRepo.published {
		t.Fatal("failed rebuild must not publish")
	}
	if !feedRepo.unlocked {
		t.Fatal("expected rebuild lock released")
	}
	// 1 次开工前清理 + 1 次失败后清理
	if feedRepo.resets != 2 {
		t.Fatalf("reset calls = %d, want 2 (initial + cleanup)", feedRepo.resets)
	}
}

// TestRebuildHotRanking_WindowScansSince 窗口模式只重建窗口内视频，并把 since 透传给仓储。
func TestRebuildHotRanking_WindowScansSince(t *testing.T) {
	cutoff := time.Now().Add(-24 * time.Hour)
	rows := []types.VideoStatSnapshot{
		{VideoID: 1, PublishAt: cutoff.Add(-time.Hour)},
		{VideoID: 2, PublishAt: cutoff.Add(time.Minute)},
		{VideoID: 3, PublishAt: cutoff.Add(2 * time.Minute)},
	}
	popular := &fakePopularRepo{rows: rows}
	feedRepo := &fakeFeedRepo{members: 2}
	svc := newRebuildTestService(popular, feedRepo)

	scanned, _, err := svc.RebuildHotRanking(context.Background(), HotRebuildOptions{Since: cutoff, BatchSize: 10})
	if err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}
	if scanned != 2 {
		t.Fatalf("scanned = %d, want 2 (only in-window rows)", scanned)
	}
	if !popular.lastSince.Equal(cutoff) {
		t.Fatalf("since = %s, want %s", popular.lastSince, cutoff)
	}
	if len(feedRepo.appended) != 2 {
		t.Fatalf("appended = %d, want 2", len(feedRepo.appended))
	}
}

package video_rank

import (
	"context"
	"testing"
	"time"

	videobasetable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_baseinfo"
	videostattable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_stat"
	"go_zero-tiktok/pkg/testutil"

	"gorm.io/gorm"
)

// seedVideo 写入一条视频；createdAt 为零值时由 GORM 填当前时间，deleted 非空表示软删。
func seedVideo(t *testing.T, db *gorm.DB, videoID int64, createdAt time.Time, deleted *time.Time) {
	t.Helper()

	video := &videobasetable.VideoBaseinfo{
		VideoID:        videoID,
		AuthorID:       1,
		VideoObjectKey: "videos/1/video.mp4",
		Title:          "t",
		CreatedAt:      createdAt,
		DeletedAt:      deleted,
	}
	testutil.AssertNoErr(t, videobasetable.CreateVideo(context.Background(), db, video))
}

// seedStat 写入一条统计行。
func seedStat(t *testing.T, db *gorm.DB, videoID, visit, like, comment, favorite int64) {
	t.Helper()

	testutil.AssertNoErr(t, videostattable.CreatePopularVideo(context.Background(), db, videoID))
	testutil.AssertNoErr(t, db.Exec(
		"UPDATE video_stat SET visit_count = ?, like_count = ?, comment_count = ?, favorite_count = ? WHERE video_id = ?",
		visit, like, comment, favorite, videoID,
	).Error)
}

// TestListCandidates_FullScanUsesIDCursor 全量模式按 video_id 游标翻页，计数来自 video_stat。
func TestListCandidates_FullScanUsesIDCursor(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	for i := int64(1); i <= 4; i++ {
		seedVideo(t, db, i, time.Time{}, nil)
	}
	// video 3 缺统计行，应被跳过
	seedStat(t, db, 1, 10, 1, 2, 3)
	seedStat(t, db, 2, 20, 2, 3, 4)
	seedStat(t, db, 4, 40, 4, 5, 6)

	page1, err := ListCandidates(ctx, db, time.Time{}, time.Time{}, 0, 2)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, len(page1), 2)
	testutil.AssertEqual(t, page1[0].VideoID, int64(1))
	testutil.AssertEqual(t, page1[0].VisitCount, int64(10))
	testutil.AssertEqual(t, page1[1].VideoID, int64(2))

	last := page1[len(page1)-1]
	page2, err := ListCandidates(ctx, db, time.Time{}, last.PublishAt, last.VideoID, 2)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, len(page2), 1)
	testutil.AssertEqual(t, page2[0].VideoID, int64(4))
	testutil.AssertEqual(t, page2[0].FavoriteCount, int64(6))

	last = page2[len(page2)-1]
	page3, err := ListCandidates(ctx, db, time.Time{}, last.PublishAt, last.VideoID, 2)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, len(page3), 0)
}

// TestListCandidates_SkipsBatchWithoutStats 整批都缺统计行时继续向后扫描，不会让调用方误判没有候选。
func TestListCandidates_SkipsBatchWithoutStats(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedVideo(t, db, 1, time.Time{}, nil)
	seedVideo(t, db, 2, time.Time{}, nil)
	seedStat(t, db, 2, 7, 0, 0, 0)

	// limit=1 时首批只会扫到 video 1（无统计行），必须继续扫到 video 2。
	rows, err := ListCandidates(ctx, db, time.Time{}, time.Time{}, 0, 1)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, len(rows), 1)
	testutil.AssertEqual(t, rows[0].VideoID, int64(2))
	testutil.AssertEqual(t, rows[0].VisitCount, int64(7))
}

// TestListCandidatesByIDs 按 ID 批量取快照：重复 ID 去重，软删与缺统计行的视频被跳过。
func TestListCandidatesByIDs(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	deletedAt := time.Now()
	seedVideo(t, db, 1, time.Time{}, nil)
	seedVideo(t, db, 2, time.Time{}, nil)
	seedVideo(t, db, 3, time.Time{}, &deletedAt)
	seedVideo(t, db, 4, time.Time{}, nil)
	seedStat(t, db, 1, 10, 1, 0, 0)
	seedStat(t, db, 2, 20, 2, 0, 0)
	seedStat(t, db, 3, 30, 3, 0, 0)
	// video 4 缺统计行

	rows, err := ListCandidatesByIDs(ctx, db, []int64{1, 2, 3, 4, 1, 0, -1})
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, len(rows), 2)

	byID := make(map[int64]CandidateRow, len(rows))
	for _, row := range rows {
		byID[row.VideoID] = row
	}
	testutil.AssertEqual(t, byID[1].VisitCount, int64(10))
	testutil.AssertEqual(t, byID[2].LikeCount, int64(2))
	if _, ok := byID[3]; ok {
		t.Fatal("soft-deleted video should be excluded")
	}
	if _, ok := byID[4]; ok {
		t.Fatal("video without stat row should be excluded")
	}
}

// TestListCandidates_WindowUsesCreatedAtCursor 窗口模式按 (created_at, video_id) 过滤，
// 首页游标 (since, 0) 包含 created_at 恰好等于 since 的视频。
func TestListCandidates_WindowUsesCreatedAtCursor(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	cutoff := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
	seedVideo(t, db, 1, cutoff.Add(-time.Hour), nil)  // 窗口外
	seedVideo(t, db, 2, cutoff, nil)                  // 边界命中
	seedVideo(t, db, 3, cutoff.Add(time.Minute), nil) // 窗口内
	seedStat(t, db, 1, 1, 0, 0, 0)
	seedStat(t, db, 2, 2, 0, 0, 0)
	seedStat(t, db, 3, 3, 0, 0, 0)

	rows, err := ListCandidates(ctx, db, cutoff, cutoff, 0, 10)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, len(rows), 2)
	testutil.AssertEqual(t, rows[0].VideoID, int64(2))
	testutil.AssertEqual(t, rows[1].VideoID, int64(3))

	total, err := CountCandidates(ctx, db, cutoff)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, total, int64(2))

	all, err := CountCandidates(ctx, db, time.Time{})
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, all, int64(3))
}

// TestListCandidates_ExcludesDeleted 软删视频不进入候选。
func TestListCandidates_ExcludesDeleted(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	deletedAt := time.Now()
	seedVideo(t, db, 1, time.Time{}, &deletedAt)
	seedVideo(t, db, 2, time.Time{}, nil)
	seedStat(t, db, 1, 1, 0, 0, 0)
	seedStat(t, db, 2, 2, 0, 0, 0)

	rows, err := ListCandidates(ctx, db, time.Time{}, time.Time{}, 0, 10)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, len(rows), 1)
	testutil.AssertEqual(t, rows[0].VideoID, int64(2))

	total, err := CountCandidates(ctx, db, time.Time{})
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, total, int64(1))
}

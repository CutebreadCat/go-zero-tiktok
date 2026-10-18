package domain

import (
	"context"
	"testing"

	videobasetable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_baseinfo"
	videostattable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_stat"
	"go_zero-tiktok/app/video/rpc/internal/dal/reposity"
	feedpkg "go_zero-tiktok/app/video/rpc/internal/domain/feed"
	"go_zero-tiktok/pkg/testutil"

	"gorm.io/gorm"
)

// newTestVideoService 用真实 DAL（临时 SQLite）构造服务，Feed/热度依赖置空以聚焦发布主链路。
func newTestVideoService(t *testing.T, db *gorm.DB) *VideoService {
	t.Helper()

	videoRepo, err := reposity.NewVideoBaseinfoRepo(db)
	testutil.AssertNoErr(t, err)
	statRepo := reposity.NewVideoStatRepo(db)

	return NewVideoService(videoRepo, statRepo, nil, nil, nil, nil, nil, nil, feedpkg.RecommendConfig{})
}

// TestPublishVideo_IdempotentByKey 客户端超时重试：即使生成了新的 videoID，
// 同一幂等键也只产生一条视频与一条统计，且返回首次的 video_id。
func TestPublishVideo_IdempotentByKey(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	svc := newTestVideoService(t, db)

	id1, at1, err := svc.PublishVideo(ctx, 1001, 7, "videos/7/1001/a.mp4", "", "标题", "描述", "key-1")
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, id1, int64(1001))
	if at1.IsZero() {
		t.Fatal("first publish should return non-zero publishAt")
	}

	// 重试：新的雪花 ID + 同一幂等键
	id2, at2, err := svc.PublishVideo(ctx, 2002, 7, "videos/7/2002/a.mp4", "", "标题", "描述", "key-1")
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, id2, int64(1001))
	if at2.IsZero() {
		t.Fatal("idempotent hit should return non-zero publishAt")
	}

	var videoCount, statCount int64
	_ = db.Model(&videobasetable.VideoBaseinfo{}).Count(&videoCount)
	_ = db.Model(&videostattable.VideoStat{}).Count(&statCount)
	testutil.AssertEqual(t, videoCount, int64(1))
	testutil.AssertEqual(t, statCount, int64(1))

	// 重试不应覆盖首次发布的内容
	var video videobasetable.VideoBaseinfo
	_ = db.Where("video_id = ?", 1001).First(&video)
	testutil.AssertEqual(t, video.VideoObjectKey, "videos/7/1001/a.mp4")
}

// TestFindPublishedVideoByKey_HealsStatRow 首次发布只成功了一半（统计行缺失）时，
// 重试命中幂等键应补齐统计行，而不是返回一个不可点赞的半成品。
func TestFindPublishedVideoByKey_HealsStatRow(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	svc := newTestVideoService(t, db)

	_, _, err := svc.PublishVideo(ctx, 1001, 7, "videos/7/1001/a.mp4", "", "标题", "描述", "key-1")
	testutil.AssertNoErr(t, err)

	// 模拟统计行写入失败/丢失
	if err := db.Where("video_id = ?", 1001).Delete(&videostattable.VideoStat{}).Error; err != nil {
		t.Fatalf("delete stat row failed: %v", err)
	}

	videoID, _, found, err := svc.FindPublishedVideoByKey(ctx, 7, "key-1")
	testutil.AssertNoErr(t, err)
	if !found {
		t.Fatal("expected idempotency key hit")
	}
	testutil.AssertEqual(t, videoID, int64(1001))

	var statCount int64
	_ = db.Model(&videostattable.VideoStat{}).Count(&statCount)
	testutil.AssertEqual(t, statCount, int64(1))
}

// TestFindPublishedVideoByKey_Miss 未命中幂等键时返回 found=false，由调用方走正常发布流程。
func TestFindPublishedVideoByKey_Miss(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestVideoService(t, db)

	_, _, found, err := svc.FindPublishedVideoByKey(context.Background(), 7, "unknown-key")
	testutil.AssertNoErr(t, err)
	if found {
		t.Fatal("unexpected idempotency key hit")
	}
}

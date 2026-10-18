package video_stat

import (
	"context"
	"testing"

	"go_zero-tiktok/pkg/testutil"
)

// TestCreatePopularVideo_Idempotency 发布重试场景：
// 基础信息已写入、统计行重复创建时应幂等成功，且只保留一行。
func TestCreatePopularVideo_Idempotency(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	testutil.AssertNoErr(t, CreatePopularVideo(ctx, db, 1))
	testutil.AssertNoErr(t, CreatePopularVideo(ctx, db, 1))

	var count int64
	_ = db.Model(&VideoStat{}).Count(&count)
	testutil.AssertEqual(t, count, int64(1))
}

// TestEnsurePopularVideo_Idempotency 消费端补充统计行的路径同样幂等。
func TestEnsurePopularVideo_Idempotency(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	testutil.AssertNoErr(t, EnsurePopularVideo(ctx, db, 2))
	testutil.AssertNoErr(t, EnsurePopularVideo(ctx, db, 2))

	var count int64
	_ = db.Model(&VideoStat{}).Count(&count)
	testutil.AssertEqual(t, count, int64(1))
}

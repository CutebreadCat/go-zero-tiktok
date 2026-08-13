package interaction

import (
	"context"
	"testing"

	"go_zero-tiktok/app/video/rpc/internal/dal/reposity"
	"go_zero-tiktok/app/video/rpc/internal/dal/tables/video_stat"
	"go_zero-tiktok/testhelpers"
)

func TestInteractionService_LikeVideo_SyncPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	// 初始化 video_stat 记录
	testhelpers.AssertNoErr(t, video_stat.CreatePopularVideo(ctx, db, 2001))

	interactionRepo := reposity.NewVideoInteractionRepo(db)
	statRepo := reposity.NewVideoStatRepo(db)
	svc := NewInteractionService(interactionRepo, statRepo, nil, nil)

	// 点赞
	testhelpers.AssertNoErr(t, svc.LikeVideo(ctx, 1001, 2001))

	// 验证 like_count = 1
	counts, err := svc.GetLikeCounts(ctx, []int64{2001})
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, counts[2001], int64(1))

	// 重复点赞应报错
	err = svc.LikeVideo(ctx, 1001, 2001)
	testhelpers.AssertInvalidParam(t, err)

	// 验证 like_count 仍为 1
	counts, err = svc.GetLikeCounts(ctx, []int64{2001})
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, counts[2001], int64(1))
}

func TestInteractionService_CancelLikeVideo_SyncPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, video_stat.CreatePopularVideo(ctx, db, 2001))

	interactionRepo := reposity.NewVideoInteractionRepo(db)
	statRepo := reposity.NewVideoStatRepo(db)
	svc := NewInteractionService(interactionRepo, statRepo, nil, nil)

	testhelpers.AssertNoErr(t, svc.LikeVideo(ctx, 1001, 2001))
	testhelpers.AssertNoErr(t, svc.CancelLikeVideo(ctx, 1001, 2001))

	counts, err := svc.GetLikeCounts(ctx, []int64{2001})
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, counts[2001], int64(0))

	// 未点赞取消应报错
	err = svc.CancelLikeVideo(ctx, 1001, 2001)
	testhelpers.AssertInvalidParam(t, err)
}

func TestInteractionService_FavoriteVideo_SyncPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, video_stat.CreatePopularVideo(ctx, db, 2001))

	interactionRepo := reposity.NewVideoInteractionRepo(db)
	statRepo := reposity.NewVideoStatRepo(db)
	svc := NewInteractionService(interactionRepo, statRepo, nil, nil)

	testhelpers.AssertNoErr(t, svc.FavoriteVideo(ctx, 1001, 2001))

	counts, err := statRepo.GetLikeCounts(ctx, []int64{2001})
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, counts[2001], int64(0))

	// 验证 favorite_count = 1
	var stat video_stat.VideoStat
	testhelpers.AssertNoErr(t, db.WithContext(ctx).First(&stat, 2001).Error)
	testhelpers.AssertEqual(t, stat.FavoriteCount, int64(1))

	// 取消收藏
	testhelpers.AssertNoErr(t, svc.CancelFavoriteVideo(ctx, 1001, 2001))
	testhelpers.AssertNoErr(t, db.WithContext(ctx).First(&stat, 2001).Error)
	testhelpers.AssertEqual(t, stat.FavoriteCount, int64(0))
}

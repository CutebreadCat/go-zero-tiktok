package interaction

import (
	"context"
	"testing"

	"go_zero-tiktok/app/interaction/rpc/internal/dal/reposity"
	"go_zero-tiktok/app/interaction/rpc/internal/dal/tables/video_stat"
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

func TestInteractionService_BatchGetVideoInteractionStats(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	// 准备两条视频及其统计
	testhelpers.AssertNoErr(t, video_stat.CreatePopularVideo(ctx, db, 2001))
	testhelpers.AssertNoErr(t, video_stat.CreatePopularVideo(ctx, db, 2002))

	interactionRepo := reposity.NewVideoInteractionRepo(db)
	statRepo := reposity.NewVideoStatRepo(db)
	svc := NewInteractionService(interactionRepo, statRepo, nil, nil)

	// 构造互动数据
	testhelpers.AssertNoErr(t, svc.LikeVideo(ctx, 1001, 2001))
	testhelpers.AssertNoErr(t, svc.FavoriteVideo(ctx, 1001, 2001))
	testhelpers.AssertNoErr(t, svc.LikeVideo(ctx, 1002, 2001))

	// 未登录场景：返回计数，liked/favorited 全 false
	stats, err := svc.BatchGetVideoInteractionStats(ctx, 0, []int64{2001, 2002, 2003})
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, stats[2001].LikeCount, int64(2))
	testhelpers.AssertEqual(t, stats[2001].FavoriteCount, int64(1))
	testhelpers.AssertEqual(t, stats[2001].Liked, false)
	testhelpers.AssertEqual(t, stats[2001].Favorited, false)
	testhelpers.AssertEqual(t, stats[2002].LikeCount, int64(0))
	testhelpers.AssertEqual(t, stats[2003].LikeCount, int64(0))
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

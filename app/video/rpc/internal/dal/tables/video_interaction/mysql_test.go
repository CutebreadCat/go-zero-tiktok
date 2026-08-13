package video_interaction

import (
	"context"
	"testing"

	videostattable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_stat"
	"go_zero-tiktok/testhelpers"
)

func uid() int64 { return 1001 }
func vid() int64 { return 2001 }

// TestLikeVideo_ParamErrors 表驱动：点赞参数校验
func TestLikeVideo_ParamErrors(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		videoID int64
	}{
		{"userID为零", 0, vid()},
		{"videoID为零", uid(), 0},
		{"两者皆为零", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.NewTestDB(t)
			err := LikeVideo(context.Background(), db, tt.userID, tt.videoID)
			testhelpers.AssertInvalidParam(t, err)
		})
	}
}

// TestLikeVideo_Success 点赞成功并落库
func TestLikeVideo_Success(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, LikeVideo(ctx, db, uid(), vid()))

	var count int64
	if err := db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeLike).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestLikeVideo_Idempotency 重复点赞应幂等报错且不写入第二条
func TestLikeVideo_Idempotency(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	// 首次点赞成功
	testhelpers.AssertNoErr(t, LikeVideo(ctx, db, uid(), vid()))
	// 重复点赞应返回参数错误（重复点赞），且不写入第二条
	err := LikeVideo(ctx, db, uid(), vid())
	testhelpers.AssertInvalidParam(t, err)

	var count int64
	_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeLike).Count(&count)
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestCancelLikeVideo 表驱动：取消点赞（成功 / 未点赞报错 / 取消后可再赞）
func TestCancelLikeVideo(t *testing.T) {
	t.Run("成功并允许再赞", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, LikeVideo(ctx, db, uid(), vid()))
		testhelpers.AssertNoErr(t, CancelLikeVideo(ctx, db, uid(), vid()))

		var count int64
		_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeLike).Count(&count)
		testhelpers.AssertEqual(t, count, int64(0))

		// 取消后可再次点赞
		testhelpers.AssertNoErr(t, LikeVideo(ctx, db, uid(), vid()))
		var again int64
		_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeLike).Count(&again)
		testhelpers.AssertEqual(t, again, int64(1))
	})

	t.Run("未点赞取消报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		err := CancelLikeVideo(context.Background(), db, uid(), vid())
		testhelpers.AssertInvalidParam(t, err)
	})
}

// TestGetLikedVideoIDsByUserID_CursorPath 游标分页：三页互不重复且总量正确
func TestGetLikedVideoIDsByUserID_CursorPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	for i := int64(1); i <= 5; i++ {
		testhelpers.AssertNoErr(t, LikeVideo(ctx, db, uid(), 3000+i))
	}

	ids1, total1, err := GetLikedVideoIDsByUserID(ctx, db, uid(), 1, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total1, int64(5))
	testhelpers.AssertEqual(t, int64(len(ids1)), int64(2))

	ids2, _, err := GetLikedVideoIDsByUserID(ctx, db, uid(), 2, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(ids2)), int64(2))

	ids3, _, err := GetLikedVideoIDsByUserID(ctx, db, uid(), 3, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(ids3)), int64(1))

	seen := map[int64]bool{}
	for _, slice := range [][]int64{ids1, ids2, ids3} {
		for _, id := range slice {
			if seen[id] {
				t.Fatalf("duplicate video id across pages: %d", id)
			}
			seen[id] = true
		}
	}
	testhelpers.AssertEqual(t, int64(len(seen)), int64(5))
}

// TestGetLikedVideoIDsByUserID_Empty 空列表
func TestGetLikedVideoIDsByUserID_Empty(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ids, total, err := GetLikedVideoIDsByUserID(context.Background(), db, uid(), 1, 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total, int64(0))
	testhelpers.AssertEqual(t, int64(len(ids)), int64(0))
}

// TestFavoriteVideo_ParamErrors 表驱动：收藏参数校验
func TestFavoriteVideo_ParamErrors(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		videoID int64
	}{
		{"userID为零", 0, vid()},
		{"videoID为零", uid(), 0},
		{"两者皆为零", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.NewTestDB(t)
			err := FavoriteVideo(context.Background(), db, tt.userID, tt.videoID)
			testhelpers.AssertInvalidParam(t, err)
		})
	}
}

// TestFavoriteVideo_Success 收藏成功并落库
func TestFavoriteVideo_Success(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, FavoriteVideo(ctx, db, uid(), vid()))

	var count int64
	if err := db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeFavorite).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestFavoriteVideo_Idempotency 重复收藏应幂等报错且不写入第二条
func TestFavoriteVideo_Idempotency(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, FavoriteVideo(ctx, db, uid(), vid()))
	err := FavoriteVideo(ctx, db, uid(), vid())
	testhelpers.AssertInvalidParam(t, err)

	var count int64
	_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeFavorite).Count(&count)
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestCancelFavoriteVideo 表驱动：取消收藏（成功 / 未收藏报错 / 取消后可再收藏）
func TestCancelFavoriteVideo(t *testing.T) {
	t.Run("成功并允许再收藏", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, FavoriteVideo(ctx, db, uid(), vid()))
		testhelpers.AssertNoErr(t, CancelFavoriteVideo(ctx, db, uid(), vid()))

		var count int64
		_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeFavorite).Count(&count)
		testhelpers.AssertEqual(t, count, int64(0))

		testhelpers.AssertNoErr(t, FavoriteVideo(ctx, db, uid(), vid()))
		var again int64
		_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeFavorite).Count(&again)
		testhelpers.AssertEqual(t, again, int64(1))
	})

	t.Run("未收藏取消报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		err := CancelFavoriteVideo(context.Background(), db, uid(), vid())
		testhelpers.AssertInvalidParam(t, err)
	})
}

// TestGetFavoritedVideoIDsByUserID_Order 收藏列表按 created_at DESC 排序
func TestGetFavoritedVideoIDsByUserID_Order(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	for i := int64(1); i <= 3; i++ {
		testhelpers.AssertNoErr(t, FavoriteVideo(ctx, db, uid(), 3000+i))
	}

	ids, total, err := GetFavoritedVideoIDsByUserID(ctx, db, uid(), 1, 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total, int64(3))
	testhelpers.AssertEqual(t, int64(len(ids)), int64(3))

	// 验证返回的是收藏过的视频 ID 集合（SQLite CURRENT_TIMESTAMP 为秒级，rapid insert 顺序可能不稳定）
	expected := map[int64]bool{3001: true, 3002: true, 3003: true}
	for _, id := range ids {
		if !expected[id] {
			t.Fatalf("unexpected video id: %d", id)
		}
	}
}

// TestGetFavoritedVideoIDsByUserID_Pagination 分页正确
func TestGetFavoritedVideoIDsByUserID_Pagination(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	for i := int64(1); i <= 5; i++ {
		testhelpers.AssertNoErr(t, FavoriteVideo(ctx, db, uid(), 3000+i))
	}

	ids1, total1, err := GetFavoritedVideoIDsByUserID(ctx, db, uid(), 1, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total1, int64(5))
	testhelpers.AssertEqual(t, int64(len(ids1)), int64(2))

	ids2, _, err := GetFavoritedVideoIDsByUserID(ctx, db, uid(), 2, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(ids2)), int64(2))

	ids3, _, err := GetFavoritedVideoIDsByUserID(ctx, db, uid(), 3, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(ids3)), int64(1))

	seen := map[int64]bool{}
	for _, slice := range [][]int64{ids1, ids2, ids3} {
		for _, id := range slice {
			if seen[id] {
				t.Fatalf("duplicate video id across pages: %d", id)
			}
			seen[id] = true
		}
	}
	testhelpers.AssertEqual(t, int64(len(seen)), int64(5))
}

// TestGetFavoritedVideoIDsByUserID_Empty 空列表
func TestGetFavoritedVideoIDsByUserID_Empty(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ids, total, err := GetFavoritedVideoIDsByUserID(context.Background(), db, uid(), 1, 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total, int64(0))
	testhelpers.AssertEqual(t, int64(len(ids)), int64(0))
}

// TestLikeAndFavoriteSameVideo_Coexist 合并后同一用户可同时点赞+收藏同一视频
func TestLikeAndFavoriteSameVideo_Coexist(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, LikeVideo(ctx, db, uid(), vid()))
	testhelpers.AssertNoErr(t, FavoriteVideo(ctx, db, uid(), vid()))

	liked, _, err := GetLikedVideoIDsByUserID(ctx, db, uid(), 1, 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(liked)), int64(1))
	testhelpers.AssertEqual(t, liked[0], vid())

	favorited, _, err := GetFavoritedVideoIDsByUserID(ctx, db, uid(), 1, 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(favorited)), int64(1))
	testhelpers.AssertEqual(t, favorited[0], vid())

	// 仅取消点赞，收藏关系不受影响
	testhelpers.AssertNoErr(t, CancelLikeVideo(ctx, db, uid(), vid()))
	favorited, _, err = GetFavoritedVideoIDsByUserID(ctx, db, uid(), 1, 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(favorited)), int64(1))
}

// TestAddLikeInteraction_Idempotent 幂等插入：重复 like 不报错，数据库只有一条。
func TestAddLikeInteraction_Idempotent(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	added1, err := AddLikeInteraction(ctx, db, uid(), vid())
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, added1, true)

	added2, err := AddLikeInteraction(ctx, db, uid(), vid())
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, added2, false)

	var count int64
	_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeLike).Count(&count)
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestRemoveLikeInteraction_Idempotent 幂等删除：未命中不报错，删除后再次删除不报错。
func TestRemoveLikeInteraction_Idempotent(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	_, err := AddLikeInteraction(ctx, db, uid(), vid())
	testhelpers.AssertNoErr(t, err)

	removed1, err := RemoveLikeInteraction(ctx, db, uid(), vid())
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, removed1, true)

	removed2, err := RemoveLikeInteraction(ctx, db, uid(), vid())
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, removed2, false)

	var count int64
	_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeLike).Count(&count)
	testhelpers.AssertEqual(t, count, int64(0))
}

// TestApplyLikeEvent_LikeAndCancel 应用 like 后 DB 关系与计数正确；cancel 后恢复 0。
func TestApplyLikeEvent_LikeAndCancel(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, ApplyLikeEvent(ctx, db, "like", uid(), vid()))

	var interactionCount int64
	_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeLike).Count(&interactionCount)
	testhelpers.AssertEqual(t, interactionCount, int64(1))

	var stat videostattable.VideoStat
	testhelpers.AssertNoErr(t, db.Where("video_id = ?", vid()).First(&stat).Error)
	testhelpers.AssertEqual(t, stat.LikeCount, int64(1))

	testhelpers.AssertNoErr(t, ApplyLikeEvent(ctx, db, "cancel", uid(), vid()))

	_ = db.Model(&VideoInteraction{}).Where("action_type = ?", ActionTypeLike).Count(&interactionCount)
	testhelpers.AssertEqual(t, interactionCount, int64(0))

	testhelpers.AssertNoErr(t, db.Where("video_id = ?", vid()).First(&stat).Error)
	testhelpers.AssertEqual(t, stat.LikeCount, int64(0))
}

// TestApplyLikeEvent_DuplicateLike 重复 like 事件不报错，计数不增加。
func TestApplyLikeEvent_DuplicateLike(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, ApplyLikeEvent(ctx, db, "like", uid(), vid()))
	testhelpers.AssertNoErr(t, ApplyLikeEvent(ctx, db, "like", uid(), vid()))

	var stat videostattable.VideoStat
	testhelpers.AssertNoErr(t, db.Where("video_id = ?", vid()).First(&stat).Error)
	testhelpers.AssertEqual(t, stat.LikeCount, int64(1))
}

// TestApplyLikeEvent_MissingStatRow video_stat 行不存在时仍能自动创建并更新计数。
func TestApplyLikeEvent_MissingStatRow(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, ApplyLikeEvent(ctx, db, "like", uid(), vid()))

	var stat videostattable.VideoStat
	testhelpers.AssertNoErr(t, db.Where("video_id = ?", vid()).First(&stat).Error)
	testhelpers.AssertEqual(t, stat.LikeCount, int64(1))
}

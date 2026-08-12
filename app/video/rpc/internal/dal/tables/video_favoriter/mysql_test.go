package video_favoriter

import (
	"context"
	"testing"

	"go_zero-tiktok/testhelpers"
)

func uid() int64 { return 1001 }
func vid() int64 { return 2001 }

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
	if err := db.Model(&VideoFavoriter{}).Count(&count).Error; err != nil {
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
	_ = db.Model(&VideoFavoriter{}).Count(&count)
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
		_ = db.Model(&VideoFavoriter{}).Count(&count)
		testhelpers.AssertEqual(t, count, int64(0))

		testhelpers.AssertNoErr(t, FavoriteVideo(ctx, db, uid(), vid()))
		var again int64
		_ = db.Model(&VideoFavoriter{}).Count(&again)
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

	// 验证返回的是收藏过的视频 ID 集合（SQLite CURRENT_TIMESTAMP 为秒级， rapid insert 顺序可能不稳定）
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

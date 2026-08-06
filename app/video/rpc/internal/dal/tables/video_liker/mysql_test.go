package video_liker

import (
	"context"
	"testing"

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
	if err := db.Model(&VideoLiker{}).Count(&count).Error; err != nil {
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
	_ = db.Model(&VideoLiker{}).Count(&count)
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
		_ = db.Model(&VideoLiker{}).Count(&count)
		testhelpers.AssertEqual(t, count, int64(0))

		// 取消后可再次点赞
		testhelpers.AssertNoErr(t, LikeVideo(ctx, db, uid(), vid()))
		var again int64
		_ = db.Model(&VideoLiker{}).Count(&again)
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

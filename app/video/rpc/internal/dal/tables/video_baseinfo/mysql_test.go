package video_baseinfo

import (
	"context"
	"fmt"
	"testing"

	"go_zero-tiktok/testhelpers"
)

func newVideo(videoID, authorID int64, title, desc string) *VideoBaseinfo {
	return &VideoBaseinfo{
		VideoID:        videoID,
		AuthorID:       authorID,
		VideoObjectKey: fmt.Sprintf("videos/%d/%d/video.mp4", authorID, videoID),
		CoverObjectKey: fmt.Sprintf("covers/%d/%d/cover.jpg", authorID, videoID),
		Title:          title,
		Description:    desc,
	}
}

// TestCreateVideo_Idempotency 相同 video_id 重复发布应幂等成功并保留首条内容
func TestCreateVideo_Idempotency(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(1, 100, "标题A", "描述")))
	err := CreateVideo(ctx, db, newVideo(1, 999, "另一个标题", "另一个描述"))
	testhelpers.AssertNoErr(t, err)

	var count int64
	_ = db.Model(&VideoBaseinfo{}).Count(&count)
	testhelpers.AssertEqual(t, count, int64(1))

	var v VideoBaseinfo
	_ = db.Where("video_id = ?", 1).First(&v)
	testhelpers.AssertEqual(t, v.Title, "标题A")
	testhelpers.AssertEqual(t, v.AuthorID, int64(100))
}

// TestGetVideosByIDs 表驱动：按 ID 批量查询（空集合 / 顺序与集合一致性）
func TestGetVideosByIDs(t *testing.T) {
	t.Run("空集合返回空", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		videos, err := GetVideosByIDs(context.Background(), db, []int64{})
		testhelpers.AssertNoErr(t, err)
		testhelpers.AssertEqual(t, int64(len(videos)), int64(0))
	})

	t.Run("按传入顺序返回", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(1, 100, "A", "")))
		testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(2, 100, "B", "")))
		testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(3, 100, "C", "")))

		videos, err := GetVideosByIDs(ctx, db, []int64{3, 1, 2})
		testhelpers.AssertNoErr(t, err)
		testhelpers.AssertEqual(t, int64(len(videos)), int64(3))
		if videos[0].VideoID != 3 || videos[1].VideoID != 1 || videos[2].VideoID != 2 {
			t.Fatalf("order not preserved: %v", []int64{videos[0].VideoID, videos[1].VideoID, videos[2].VideoID})
		}
	})
}

// TestGetVideosByAuthorID_CursorPath 作者视频列表游标分页
func TestGetVideosByAuthorID_CursorPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()
	for i := int64(1); i <= 5; i++ {
		testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(100+i, 500, "t", "")))
	}

	page1, total, err := GetVideosByAuthorID(ctx, db, 500, 1, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total, int64(5))
	testhelpers.AssertEqual(t, int64(len(page1)), int64(2))

	page2, _, err := GetVideosByAuthorID(ctx, db, 500, 2, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(page2)), int64(2))

	page3, _, err := GetVideosByAuthorID(ctx, db, 500, 3, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(page3)), int64(1))
}

// TestSearchVideosByKeyword 表驱动：关键词搜索（匹配 / 空关键词返回全部）
func TestSearchVideosByKeyword(t *testing.T) {
	tests := []struct {
		name      string
		keyword   string
		wantTotal int64
	}{
		{"匹配小猫", "小猫", 2},
		{"空关键词返回全部", "", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.NewTestDB(t)
			ctx := context.Background()
			testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(1, 100, "小猫钓鱼", "可爱")))
			testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(2, 100, "小狗跑步", "快乐")))
			testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(3, 100, "小猫睡觉", "安静")))

			videos, total, err := SearchVideosByKeyword(ctx, db, tt.keyword, 1, 10)
			testhelpers.AssertNoErr(t, err)
			testhelpers.AssertEqual(t, total, tt.wantTotal)
			testhelpers.AssertEqual(t, int64(len(videos)), tt.wantTotal)
		})
	}
}

// TestGetVideoByLastTime_CursorPath 时间游标分页与非法时间
func TestGetVideoByLastTime_CursorPath(t *testing.T) {
	t.Run("游标翻页", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		for i := int64(1); i <= 5; i++ {
			testhelpers.AssertNoErr(t, CreateVideo(ctx, db, newVideo(100+i, 500, "t", "")))
		}

		page1, total, err := GetVideoByLastTime(ctx, db, "", 1, 2)
		testhelpers.AssertNoErr(t, err)
		testhelpers.AssertEqual(t, total, int64(5))
		testhelpers.AssertEqual(t, int64(len(page1)), int64(2))

		last := page1[len(page1)-1]
		page2, _, err := GetVideoByLastTime(ctx, db, last.CreatedAt.Format("2006-01-02 15:04:05.000"), 1, 2)
		testhelpers.AssertNoErr(t, err)
		if int64(len(page2)) == 0 {
			t.Fatal("expected more videos via cursor")
		}
	})

	t.Run("非法时间报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		_, _, err := GetVideoByLastTime(context.Background(), db, "not-a-time", 1, 10)
		testhelpers.AssertErr(t, err)
	})
}

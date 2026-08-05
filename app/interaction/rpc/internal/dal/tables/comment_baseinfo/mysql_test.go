package comment_baseinfo

import (
	"context"
	"testing"

	"go_zero-tiktok/testhelpers"
)

// TestCreateComment_ParamErrors 表驱动：创建评论参数校验
func TestCreateComment_ParamErrors(t *testing.T) {
	tests := []struct {
		name    string
		comment *CommentBaseinfo
	}{
		{"nil评论", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.NewTestDB(t)
			err := CreateComment(context.Background(), db, tt.comment)
			testhelpers.AssertInvalidParam(t, err)
		})
	}
}

// TestCreateComment_Success 创建评论成功落库
func TestCreateComment_Success(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	comment := &CommentBaseinfo{CommentID: 1, UserID: 100, VideoID: 200, Content: "好视频"}
	testhelpers.AssertNoErr(t, CreateComment(context.Background(), db, comment))

	var count int64
	_ = db.Model(&CommentBaseinfo{}).Count(&count)
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestCreateComment_Idempotency 相同 comment_id 重复创建应幂等成功且保留首条内容
func TestCreateComment_Idempotency(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()
	testhelpers.AssertNoErr(t, CreateComment(ctx, db, &CommentBaseinfo{CommentID: 1, UserID: 100, VideoID: 200, Content: "原内容"}))
	err := CreateComment(ctx, db, &CommentBaseinfo{CommentID: 1, UserID: 999, VideoID: 999, Content: "新内容"})
	testhelpers.AssertNoErr(t, err)

	var count int64
	_ = db.Model(&CommentBaseinfo{}).Count(&count)
	testhelpers.AssertEqual(t, count, int64(1))

	var c CommentBaseinfo
	_ = db.Where("comment_id = ?", 1).First(&c)
	testhelpers.AssertEqual(t, c.Content, "原内容")
	testhelpers.AssertEqual(t, c.UserID, int64(100))
}

// TestDeleteCommentByID 表驱动：删除评论（成功 / 不存在报错 / 非作者报错）
func TestDeleteCommentByID(t *testing.T) {
	t.Run("成功删除", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, CreateComment(ctx, db, &CommentBaseinfo{CommentID: 1, UserID: 100, VideoID: 200, Content: "c"}))
		testhelpers.AssertNoErr(t, DeleteCommentByID(ctx, db, 1, 100))

		var count int64
		_ = db.Model(&CommentBaseinfo{}).Count(&count)
		testhelpers.AssertEqual(t, count, int64(0))
	})

	t.Run("评论不存在报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		err := DeleteCommentByID(context.Background(), db, 999, 100)
		testhelpers.AssertInvalidParam(t, err)
	})

	t.Run("非作者报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, CreateComment(ctx, db, &CommentBaseinfo{CommentID: 1, UserID: 100, VideoID: 200, Content: "c"}))
		err := DeleteCommentByID(ctx, db, 1, 200)
		testhelpers.AssertInvalidParam(t, err)
	})
}

// TestGetCommentsByVideoID_CursorPath 评论列表游标分页
func TestGetCommentsByVideoID_CursorPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()
	for i := int64(1); i <= 5; i++ {
		testhelpers.AssertNoErr(t, CreateComment(ctx, db, &CommentBaseinfo{CommentID: 100 + i, UserID: 100, VideoID: 200, Content: "c"}))
	}

	page1, total, err := GetCommentsByVideoID(ctx, db, 200, 1, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total, int64(5))
	testhelpers.AssertEqual(t, int64(len(page1)), int64(2))

	page2, _, err := GetCommentsByVideoID(ctx, db, 200, 2, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(page2)), int64(2))

	page3, _, err := GetCommentsByVideoID(ctx, db, 200, 3, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(page3)), int64(1))
}

// TestLikeComment 表驱动：评论点赞（成功 / 重复报错 / 参数校验 / 取消）
func TestLikeComment(t *testing.T) {
	t.Run("成功并重复报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, LikeComment(ctx, db, 1, 100))
		err := LikeComment(ctx, db, 1, 100)
		testhelpers.AssertInvalidParam(t, err)

		var count int64
		_ = db.Model(&CommentLiker{}).Count(&count)
		testhelpers.AssertEqual(t, count, int64(1))
	})

	t.Run("参数校验零值", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		err := LikeComment(context.Background(), db, 0, 100)
		testhelpers.AssertInvalidParam(t, err)
	})

	t.Run("取消点赞成功与报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, LikeComment(ctx, db, 1, 100))
		testhelpers.AssertNoErr(t, UnlikeComment(ctx, db, 1, 100))

		var count int64
		_ = db.Model(&CommentLiker{}).Count(&count)
		testhelpers.AssertEqual(t, count, int64(0))

		// 未点赞取消报错
		err := UnlikeComment(ctx, db, 1, 100)
		testhelpers.AssertInvalidParam(t, err)
	})
}

// TestCommentPareantComment_SuccessAndIdempotency 回复评论生成并返回 ID
func TestCommentPareantComment_SuccessAndIdempotency(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	id1, err := CommentPareantComment(ctx, db, 0, "回复内容", 100, 200)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, id1 != 0, true)

	id2, err := CommentPareantComment(ctx, db, 0, "另一回复", 101, 200)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, id1 != id2, true)

	var count int64
	_ = db.Model(&CommentBaseinfo{}).Count(&count)
	testhelpers.AssertEqual(t, count, int64(2))
}

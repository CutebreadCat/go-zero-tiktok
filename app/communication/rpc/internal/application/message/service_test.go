package applicationmessage

import (
	"context"
	"testing"

	"go_zero-tiktok/app/communication/rpc/internal/dal/reposity"
	"go_zero-tiktok/testhelpers"
)

// TestService_CreateAndList 应用服务创建消息并分页查询
func TestService_CreateAndList(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	svc := New(reposity.NewMessageRepo(db))
	ctx := context.Background()

	res, err := svc.Create(ctx, 100, "LIKE", "收到点赞", "有人点赞了你的视频",
		"evt:svc:1", "idem:svc:1", 200, "user200", "http://avatar", 300, "video")
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, res.Created, true)
	testhelpers.AssertEqual(t, res.Message.ReceiverID, int64(100))
	testhelpers.AssertEqual(t, res.Message.SenderNickname, "user200")

	// 幂等重复创建
	res2, err := svc.Create(ctx, 100, "LIKE", "收到点赞", "有人点赞了你的视频",
		"evt:svc:1", "idem:svc:2", 200, "user200", "http://avatar", 300, "video")
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, res2.Created, false)
	testhelpers.AssertEqual(t, res2.Message.ID, res.Message.ID)

	// 自己给自己不发消息
	res3, err := svc.Create(ctx, 100, "LIKE", "收到点赞", "有人点赞了你的视频",
		"evt:svc:self", "idem:svc:self", 100, "self", "", 300, "video")
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, res3.Created, false)
	testhelpers.AssertEqual(t, res3.Message == nil, true)

	list, err := svc.List(ctx, 100, "", "", 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, len(list.Items), 1)
	testhelpers.AssertEqual(t, list.HasMore, false)
}

// TestService_CountUnreadAndMarkRead 未读数与已读
func TestService_CountUnreadAndMarkRead(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	svc := New(reposity.NewMessageRepo(db))
	ctx := context.Background()

	for i := int64(1); i <= 3; i++ {
		_, err := svc.Create(ctx, 100, "FOLLOW", "新增关注", "有人关注了你",
			"evt:svc:follow:"+string(rune('0'+i)), "idem:svc:follow:"+string(rune('0'+i)), 200+i, "", "", 200+i, "user")
		testhelpers.AssertNoErr(t, err)
	}

	count, err := svc.CountUnread(ctx, 100)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, count, int64(3))

	list, err := svc.List(ctx, 100, "", "", 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, len(list.Items), 2)
	testhelpers.AssertEqual(t, list.HasMore, true)

	ids := []int64{list.Items[0].ID, list.Items[1].ID}
	affected, err := svc.MarkRead(ctx, 100, ids)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, affected, int64(2))

	count, err = svc.CountUnread(ctx, 100)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestService_InvalidType 非法消息类型应报错
func TestService_InvalidType(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	svc := New(reposity.NewMessageRepo(db))
	ctx := context.Background()

	_, err := svc.Create(ctx, 100, "UNKNOWN", "标题", "内容", "evt:bad", "idem:bad", 0, "", "", 0, "")
	testhelpers.AssertErr(t, err)
}

package reposity

import (
	"context"
	"testing"
	"time"

	domainmessage "go_zero-tiktok/app/communication/rpc/internal/domain/message"
	"go_zero-tiktok/pkg/testutil"
)

func newDomainMessage(receiverID int64, eventID string) *domainmessage.Message {
	m, _ := domainmessage.New(receiverID, "COMMENT", "收到评论", "有人评论了你的视频", eventID)
	_, _ = m.AssignID()
	m.WithSender(200, "sender", "http://avatar").WithTarget(300, "video")
	m.CreatedAt = time.Now()
	return m
}

// TestMessageRepo_CreateAndList 创建消息后能按接收人查出
func TestMessageRepo_CreateAndList(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := NewMessageRepo(db)
	ctx := context.Background()

	m := newDomainMessage(100, "evt:repo:1")
	created, inserted, err := repo.Create(ctx, m, "idem:repo:1")
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, inserted, true)
	testutil.AssertEqual(t, created.ReceiverID, int64(100))
	testutil.AssertEqual(t, created.SenderID, int64(200))
	testutil.AssertEqual(t, created.TargetType, "video")

	items, err := repo.ListByUser(ctx, 100, "", nil, 10)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, len(items), 1)
	testutil.AssertEqual(t, items[0].Content, "有人评论了你的视频")
}

// TestMessageRepo_CountUnreadAndMarkRead 未读数与标记已读
func TestMessageRepo_CountUnreadAndMarkRead(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := NewMessageRepo(db)
	ctx := context.Background()

	for i := int64(1); i <= 2; i++ {
		m := newDomainMessage(100, "evt:repo:unread:"+string(rune('0'+i)))
		_, _, err := repo.Create(ctx, m, "idem:repo:unread:"+string(rune('0'+i)))
		testutil.AssertNoErr(t, err)
	}

	count, err := repo.CountUnread(ctx, 100)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, count, int64(2))

	// 查出所有 id
	items, err := repo.ListByUser(ctx, 100, "", nil, 10)
	testutil.AssertNoErr(t, err)
	ids := make([]int64, len(items))
	for i, it := range items {
		ids[i] = it.ID
	}

	affected, err := repo.MarkRead(ctx, 100, ids)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, affected, int64(2))

	count, err = repo.CountUnread(ctx, 100)
	testutil.AssertNoErr(t, err)
	testutil.AssertEqual(t, count, int64(0))
}

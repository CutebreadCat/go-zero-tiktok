package messagestable

import (
	"context"
	"testing"
	"time"

	domainmessage "go_zero-tiktok/app/communication/rpc/internal/domain/message"
	"go_zero-tiktok/testhelpers"
)

func newMessage(id int64, receiverID int64, eventID string) *domainmessage.Message {
	m, _ := domainmessage.New(receiverID, "LIKE", "收到点赞", "有人点赞了你的视频", eventID)
	m.ID = id
	m.SenderID = 200
	m.TargetID = 300
	m.TargetType = "video"
	m.CreatedAt = time.Now()
	return m
}

// TestCreate_Success 正常创建消息并落库
func TestCreate_Success(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	m := newMessage(1001, 100, "evt:1")
	record, created, err := Create(ctx, db, m, "idem:1")
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, created, true)
	testhelpers.AssertEqual(t, record.MessageID, int64(1001))
	testhelpers.AssertEqual(t, record.ReceiverID, int64(100))
	testhelpers.AssertEqual(t, record.IsRead, int8(0))
}

// TestCreate_Idempotency 相同 event_id 重复创建应幂等返回已有记录
func TestCreate_Idempotency(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	m1 := newMessage(1001, 100, "evt:1")
	_, created1, err := Create(ctx, db, m1, "idem:1")
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, created1, true)

	m2 := newMessage(1002, 100, "evt:1")
	record2, created2, err := Create(ctx, db, m2, "idem:2")
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, created2, false)
	testhelpers.AssertEqual(t, record2.MessageID, int64(1001))
}

// TestListByUser_CursorPath 按创建时间倒序分页
func TestListByUser_CursorPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	for i := int64(1); i <= 5; i++ {
		m := newMessage(1000+i, 100, "evt:list:"+string(rune('0'+i)))
		m.CreatedAt = time.Now().Add(time.Duration(-i) * time.Second)
		_, _, err := Create(ctx, db, m, "idem:list:"+string(rune('0'+i)))
		testhelpers.AssertNoErr(t, err)
	}

	page1, err := ListByUser(ctx, db, 100, "", nil, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, len(page1), 2)
	testhelpers.AssertEqual(t, page1[0].MessageID, int64(1001))
	testhelpers.AssertEqual(t, page1[1].MessageID, int64(1002))

	cursor := &domainmessage.Cursor{CreatedAt: page1[1].CreatedAt, MessageID: page1[1].ID}
	page2, err := ListByUser(ctx, db, 100, "", cursor, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, len(page2), 2)
}

// TestCountUnread_And_MarkRead 未读数统计与标记已读
func TestCountUnread_And_MarkRead(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	for i := int64(1); i <= 3; i++ {
		m := newMessage(2000+i, 100, "evt:unread:"+string(rune('0'+i)))
		_, _, err := Create(ctx, db, m, "idem:unread:"+string(rune('0'+i)))
		testhelpers.AssertNoErr(t, err)
	}

	count, err := CountUnread(ctx, db, 100)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, count, int64(3))

	// 标记前两条已读
	var ids []int64
	_ = db.WithContext(ctx).Model(&Message{}).Where("receiver_id = ?", 100).Limit(2).Pluck("id", &ids)
	affected, err := MarkRead(ctx, db, 100, ids)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, affected, int64(2))

	count, err = CountUnread(ctx, db, 100)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, count, int64(1))
}

package interaction

import (
	"context"
	"testing"

	"go_zero-tiktok/app/interaction/rpc/internal/dal/reposity"
	videointeractiontable "go_zero-tiktok/app/interaction/rpc/internal/dal/tables/video_interaction"
	videostattable "go_zero-tiktok/app/interaction/rpc/internal/dal/tables/video_stat"
	"go_zero-tiktok/testhelpers"
)

func TestLikeEventHandler_Consume_Like(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	repo := reposity.NewVideoInteractionRepo(db)
	handler := NewLikeEventHandler(repo)
	ctx := context.Background()

	event := (&LikeEvent{
		UserID:  1001,
		VideoID: 2001,
		Action:  LikeActionLike,
	}).ToKafkaEvent(DefaultLikeTopic)

	testhelpers.AssertNoErr(t, handler.Consume(ctx, event))

	var interactionCount int64
	_ = db.Model(&videointeractiontable.VideoInteraction{}).
		Where("user_id = ? AND video_id = ? AND action_type = ?", 1001, 2001, videointeractiontable.ActionTypeLike).
		Count(&interactionCount)
	testhelpers.AssertEqual(t, interactionCount, int64(1))

	var stat videostattable.VideoStat
	testhelpers.AssertNoErr(t, db.Where("video_id = ?", 2001).First(&stat).Error)
	testhelpers.AssertEqual(t, stat.LikeCount, int64(1))
}

func TestLikeEventHandler_Consume_Cancel(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	repo := reposity.NewVideoInteractionRepo(db)
	handler := NewLikeEventHandler(repo)
	ctx := context.Background()

	// 先点赞
	testhelpers.AssertNoErr(t, repo.ApplyLikeEvent(ctx, "like", 1001, 2001))

	// 消费取消事件
	event := (&LikeEvent{
		UserID:  1001,
		VideoID: 2001,
		Action:  LikeActionCancel,
	}).ToKafkaEvent(DefaultLikeTopic)
	testhelpers.AssertNoErr(t, handler.Consume(ctx, event))

	var interactionCount int64
	_ = db.Model(&videointeractiontable.VideoInteraction{}).
		Where("user_id = ? AND video_id = ? AND action_type = ?", 1001, 2001, videointeractiontable.ActionTypeLike).
		Count(&interactionCount)
	testhelpers.AssertEqual(t, interactionCount, int64(0))

	var stat videostattable.VideoStat
	testhelpers.AssertNoErr(t, db.Where("video_id = ?", 2001).First(&stat).Error)
	testhelpers.AssertEqual(t, stat.LikeCount, int64(0))
}

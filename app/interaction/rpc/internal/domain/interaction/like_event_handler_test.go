package interaction

import (
	"context"
	"testing"

	"go_zero-tiktok/testhelpers"
)

type mockLikeDirtyMarker struct {
	likeVideos     []int64
	favoriteVideos []int64
}

func (m *mockLikeDirtyMarker) MarkVideoLikeDirty(_ context.Context, videoID int64) error {
	m.likeVideos = append(m.likeVideos, videoID)
	return nil
}

func (m *mockLikeDirtyMarker) MarkVideoFavoriteDirty(_ context.Context, videoID int64) error {
	m.favoriteVideos = append(m.favoriteVideos, videoID)
	return nil
}

func TestLikeEventHandler_Consume_Like(t *testing.T) {
	marker := &mockLikeDirtyMarker{}
	handler := NewLikeEventHandler(marker)
	ctx := context.Background()

	event := (&LikeEvent{
		UserID:  1001,
		VideoID: 2001,
		Action:  LikeActionLike,
	}).ToKafkaEvent(DefaultLikeTopic)

	testhelpers.AssertNoErr(t, handler.Consume(ctx, event))
	testhelpers.AssertEqual(t, len(marker.likeVideos), 1)
	testhelpers.AssertEqual(t, marker.likeVideos[0], int64(2001))
	testhelpers.AssertEqual(t, len(marker.favoriteVideos), 0)
}

func TestLikeEventHandler_Consume_Cancel(t *testing.T) {
	marker := &mockLikeDirtyMarker{}
	handler := NewLikeEventHandler(marker)
	ctx := context.Background()

	event := (&LikeEvent{
		UserID:  1001,
		VideoID: 2001,
		Action:  LikeActionCancel,
	}).ToKafkaEvent(DefaultLikeTopic)

	testhelpers.AssertNoErr(t, handler.Consume(ctx, event))
	testhelpers.AssertEqual(t, len(marker.likeVideos), 1)
	testhelpers.AssertEqual(t, marker.likeVideos[0], int64(2001))
}

func TestLikeEventHandler_Consume_Favorite(t *testing.T) {
	marker := &mockLikeDirtyMarker{}
	handler := NewLikeEventHandler(marker)
	ctx := context.Background()

	event := (&LikeEvent{
		UserID:  1001,
		VideoID: 2002,
		Action:  LikeActionFavorite,
	}).ToKafkaEvent(DefaultLikeTopic)

	testhelpers.AssertNoErr(t, handler.Consume(ctx, event))
	testhelpers.AssertEqual(t, len(marker.favoriteVideos), 1)
	testhelpers.AssertEqual(t, marker.favoriteVideos[0], int64(2002))
}

func TestLikeEventHandler_Consume_UnknownAction(t *testing.T) {
	marker := &mockLikeDirtyMarker{}
	handler := NewLikeEventHandler(marker)
	ctx := context.Background()

	event := (&LikeEvent{
		UserID:  1001,
		VideoID: 2003,
		Action:  "unknown",
	}).ToKafkaEvent(DefaultLikeTopic)

	testhelpers.AssertNoErr(t, handler.Consume(ctx, event))
	testhelpers.AssertEqual(t, len(marker.likeVideos), 0)
	testhelpers.AssertEqual(t, len(marker.favoriteVideos), 0)
}

package reposity

import (
	"context"

	videovieweventtable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_view_events"

	"gorm.io/gorm"
)

// VideoViewEventRepo 视频浏览/播放事件仓库。
type VideoViewEventRepo struct {
	db *gorm.DB
}

// NewVideoViewEventRepo 创建视频浏览/播放事件仓库。
func NewVideoViewEventRepo(db *gorm.DB) *VideoViewEventRepo {
	return &VideoViewEventRepo{db: db}
}

// CreateEvent 写入一条播放/浏览事件。
func (r *VideoViewEventRepo) CreateEvent(ctx context.Context, userID, videoID int64, scene, requestID, eventType string, watchMs int64, completed int8) error {
	row := &videovieweventtable.VideoViewEvent{
		UserID:    userID,
		VideoID:   videoID,
		Scene:     scene,
		RequestID: requestID,
		EventType: eventType,
		WatchMs:   watchMs,
		Completed: completed,
	}
	return videovieweventtable.CreateEvent(ctx, r.db, row)
}

// BatchCreateEvents 批量写入播放/浏览事件。
func (r *VideoViewEventRepo) BatchCreateEvents(ctx context.Context, events []*videovieweventtable.VideoViewEvent) error {
	return videovieweventtable.BatchCreateEvents(ctx, r.db, events)
}

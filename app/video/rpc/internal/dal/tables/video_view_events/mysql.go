package video_view_events

import (
	"context"

	"go_zero-tiktok/pkg/xerr"

	"gorm.io/gorm"
)

// CreateEvent 写入一条视频浏览/播放事件记录。
func CreateEvent(ctx context.Context, db *gorm.DB, event *VideoViewEvent) error {
	if event == nil {
		return xerr.NewInvalidParam("事件数据为空")
	}
	if err := db.WithContext(ctx).Create(event).Error; err != nil {
		return xerr.Wrap(err, "create video view event failed")
	}
	return nil
}

// BatchCreateEvents 批量写入视频浏览/播放事件记录。
func BatchCreateEvents(ctx context.Context, db *gorm.DB, events []*VideoViewEvent) error {
	if len(events) == 0 {
		return nil
	}
	if err := db.WithContext(ctx).CreateInBatches(events, 100).Error; err != nil {
		return xerr.Wrap(err, "batch create video view events failed")
	}
	return nil
}

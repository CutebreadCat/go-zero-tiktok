package video_view_events

import "time"

// VideoViewEvent 视频浏览/播放事件数据库模型（对齐 migrations/000005_video_view_events）。
type VideoViewEvent struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;type:bigint;column:id"`
	UserID    int64     `gorm:"not null;type:bigint;column:user_id"`
	VideoID   int64     `gorm:"not null;type:bigint;column:video_id"`
	Scene     string    `gorm:"type:varchar(32);column:scene;default:''"`
	RequestID string    `gorm:"type:varchar(64);column:request_id;default:''"`
	EventType string    `gorm:"type:varchar(32);column:event_type"`
	WatchMs   int64     `gorm:"type:bigint;column:watch_ms;default:0"`
	Completed int8      `gorm:"type:tinyint;column:completed;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
}

func (VideoViewEvent) TableName() string {
	return "video_view_events"
}

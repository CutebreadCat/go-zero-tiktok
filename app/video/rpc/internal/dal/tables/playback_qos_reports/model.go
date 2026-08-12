package playback_qos_reports

import (
	"time"
)

// PlaybackQoSReport 播放质量上报数据库模型(对齐 migrations/playback_qos_reports)
type PlaybackQoSReport struct {
	ID             int64     `gorm:"primaryKey;autoIncrement;type:bigint;column:id"`
	UserID         int64     `gorm:"not null;type:bigint;column:user_id"`
	VideoID        int64     `gorm:"not null;type:bigint;column:video_id"`
	ReportData     string    `gorm:"type:json;column:report_data"`
	IdempotencyKey string    `gorm:"type:varchar(64);column:idempotency_key"`
	CreatedAt      time.Time `gorm:"autoCreateTime;column:created_at"`
}

func (PlaybackQoSReport) TableName() string {
	return "playback_qos_reports"
}

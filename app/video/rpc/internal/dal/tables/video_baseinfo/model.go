package video_baseinfo

import (
	"time"
)

// VideoBaseinfo 视频基础信息数据库模型
type VideoBaseinfo struct {
	VideoID        int64      `gorm:"primaryKey;type:bigint;column:video_id"`
	AuthorID       int64      `gorm:"not null;type:bigint;column:author_id"`
	VideoURL       string     `gorm:"not null;type:varchar(255);column:video_url"`
	CoverURL       string     `gorm:"type:varchar(255);column:cover_url"`
	Title          string     `gorm:"not null;type:varchar(128);column:title"`
	Description    string     `gorm:"type:varchar(255);column:description"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime;column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
	IdempotencyKey *string    `gorm:"type:varchar(64);column:idempotency_key"`
}

func (VideoBaseinfo) TableName() string {
	return "video_baseinfo"
}

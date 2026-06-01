package video_baseinfo

import (
	"time"
)

// VideoBaseinfo 视频基础信息数据库模型
type VideoBaseinfo struct {
	VideoID     string     `gorm:"primaryKey;type:varchar(64);column:video_id"`
	AuthorID    string     `gorm:"not null;type:varchar(64);column:author_id"`
	VideoURL    string     `gorm:"not null;type:varchar(512);column:video_url"`
	CoverURL    string     `gorm:"type:varchar(512);column:cover_url"`
	Title       string     `gorm:"not null;type:varchar(255);column:title"`
	Description string     `gorm:"type:varchar(512);column:description"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime;column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (VideoBaseinfo) TableName() string {
	return "video_baseinfo"
}

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

// VideoPopular 视频热度数据库模型
type VideoPopular struct {
	VideoID      string `gorm:"primaryKey;type:varchar(64);column:video_id"`
	VisitCount   int64  `gorm:"default:0;type:bigint;column:visit_count"`
	LikeCount    int64  `gorm:"default:0;type:bigint;column:like_count"`
	CommentCount int64  `gorm:"default:0;type:bigint;column:comment_count"`
}

func (VideoPopular) TableName() string {
	return "video_popular"
}

// VideoLiker 视频点赞数据库模型
type VideoLiker struct {
	UserID  string `gorm:"primaryKey;type:varchar(64);column:user_id"`
	VideoID string `gorm:"primaryKey;type:varchar(64);column:video_id"`
}

func (VideoLiker) TableName() string {
	return "video_liker"
}

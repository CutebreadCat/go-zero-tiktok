package comment_baseinfo

import (
	"time"
)

// CommentBaseinfo 评论基础信息数据库模型
type CommentBaseinfo struct {
	CommentID       int64      `gorm:"primaryKey;type:bigint;column:comment_id"`
	UserID          int64      `gorm:"not null;type:bigint;column:user_id"`
	VideoID         int64      `gorm:"not null;type:bigint;column:video_id"`
	Content         string     `gorm:"not null;type:varchar(512);column:content"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime;column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
	LikeCount       int64      `gorm:"default:0;type:bigint;column:like_count"`
	ParentCommentID int64      `gorm:"default:0;type:bigint;column:parent_comment_id"`
	IdempotencyKey  *string    `gorm:"type:varchar(64);column:idempotency_key"`
}

func (CommentBaseinfo) TableName() string {
	return "comment_baseinfo"
}

// CommentLiker 评论点赞数据库模型
type CommentLiker struct {
	UserID         int64   `gorm:"primaryKey;type:bigint;column:user_id"`
	CommentID      int64   `gorm:"primaryKey;type:bigint;column:comment_id"`
	IdempotencyKey *string `gorm:"type:varchar(64);column:idempotency_key"`
}

func (CommentLiker) TableName() string {
	return "comment_liker"
}

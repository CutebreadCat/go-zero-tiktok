package comment_baseinfo

import (
	"time"
)

// CommentBaseinfo 评论基础信息数据库模型
type CommentBaseinfo struct {
	CommentID       string     `gorm:"primaryKey;type:varchar(64);column:comment_id"`
	UserID          string     `gorm:"not null;type:varchar(64);column:user_id"`
	VideoID         string     `gorm:"not null;type:varchar(64);column:video_id"`
	Content         string     `gorm:"not null;type:varchar(1024);column:content"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime;column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
	LikeCount       int32      `gorm:"default:0;type:int;column:like_count"`
	ParentCommentID string     `gorm:"type:char(64);default:'';column:parent_comment_id"`
}

func (CommentBaseinfo) TableName() string {
	return "comment_baseinfo"
}

// CommentLiker 评论点赞数据库模型
type CommentLiker struct {
	UserID    string `gorm:"primaryKey;type:varchar(64);column:user_id"`
	CommentID string `gorm:"primaryKey;type:varchar(64);column:comment_id"`
}

func (CommentLiker) TableName() string {
	return "comment_liker"
}

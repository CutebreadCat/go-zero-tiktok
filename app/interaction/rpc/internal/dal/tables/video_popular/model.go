package video_popular

import "time"

type VideoPopular struct {
	VideoID      string    `gorm:"column:video_id;primaryKey"`
	VisitCount   int64     `gorm:"column:visit_count"`
	LikeCount    int64     `gorm:"column:like_count"`
	CommentCount int64     `gorm:"column:comment_count"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (VideoPopular) TableName() string {
	return "video_popular"
}

package video_popular

// VideoPopular 视频热度数据库模型
type VideoPopular struct {
	VideoID      int64 `gorm:"primaryKey;type:bigint;column:video_id"`
	VisitCount   int64 `gorm:"default:0;type:bigint;column:visit_count"`
	LikeCount    int64 `gorm:"default:0;type:bigint;column:like_count"`
	CommentCount int64 `gorm:"default:0;type:bigint;column:comment_count"`
}

func (VideoPopular) TableName() string {
	return "video_popular"
}

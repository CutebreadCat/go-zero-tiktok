package video_stat

// VideoStat 视频统计数据库模型(原 video_popular,重构后命名对齐 migrations)
type VideoStat struct {
	VideoID      int64 `gorm:"primaryKey;type:bigint;column:video_id"`
	VisitCount   int64 `gorm:"default:0;type:bigint;column:visit_count"`
	LikeCount    int64 `gorm:"default:0;type:bigint;column:like_count"`
	CommentCount int64 `gorm:"default:0;type:bigint;column:comment_count"`
}

func (VideoStat) TableName() string {
	return "video_stat"
}
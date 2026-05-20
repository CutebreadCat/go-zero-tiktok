package video_popular

const (
	popularVideosRankKey = "popular_videos"
	popularVideosHashKey = "popular_videos:hash"
)

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

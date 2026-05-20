package video_liker

// VideoLiker 视频点赞数据库模型
type VideoLiker struct {
	UserID  string `gorm:"primaryKey;type:varchar(64);column:user_id"`
	VideoID string `gorm:"primaryKey;type:varchar(64);column:video_id"`
}

func (VideoLiker) TableName() string {
	return "video_liker"
}

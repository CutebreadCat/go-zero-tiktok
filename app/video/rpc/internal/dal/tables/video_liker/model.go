package video_liker

// VideoLiker 视频点赞数据库模型
type VideoLiker struct {
	UserID         int64   `gorm:"primaryKey;type:bigint;column:user_id"`
	VideoID        int64   `gorm:"primaryKey;type:bigint;column:video_id"`
	IdempotencyKey *string `gorm:"type:varchar(64);column:idempotency_key"`
}

func (VideoLiker) TableName() string {
	return "video_liker"
}

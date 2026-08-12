package video_favoriter

// VideoFavoriter 视频收藏数据库模型
type VideoFavoriter struct {
	UserID         int64   `gorm:"primaryKey;type:bigint;column:user_id"`
	VideoID        int64   `gorm:"primaryKey;type:bigint;column:video_id"`
	IdempotencyKey *string `gorm:"type:varchar(64);column:idempotency_key"`
}

func (VideoFavoriter) TableName() string {
	return "video_favoriter"
}

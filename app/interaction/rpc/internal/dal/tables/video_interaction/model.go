package video_interaction

// VideoInteraction 视频交互(点赞/收藏)数据库模型
// 由原 video_liker(点赞) 与 video_favoriter(收藏) 合并而来，以 action_type 区分。
type VideoInteraction struct {
	ID             int64   `gorm:"primaryKey;autoIncrement;type:bigint;column:id"`
	UserID         int64   `gorm:"type:bigint;column:user_id"`
	VideoID        int64   `gorm:"type:bigint;column:video_id"`
	ActionType     int32   `gorm:"type:tinyint;column:action_type;default:1"`
	IdempotencyKey *string `gorm:"type:varchar(64);column:idempotency_key"`
}

func (VideoInteraction) TableName() string {
	return "video_interaction"
}

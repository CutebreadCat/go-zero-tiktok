package user_follow

import "time"

// UserFollow 用户关注关系数据库模型
// 方案B: 软删除使用 deleted_at + active_flag 生成列，active_flag 由 MySQL 自动推导(只读)
type UserFollow struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;type:bigint;column:id"`
	FollowerID     int64      `gorm:"type:bigint;column:follower_id"`
	UserID         int64      `gorm:"type:bigint;column:user_id"`
	IdempotencyKey *string    `gorm:"type:varchar(64);column:idempotency_key"`
	DeletedAt      *time.Time `gorm:"type:datetime(3);column:deleted_at"`
	ActiveFlag     *int8      `gorm:"->;-:migration;column:active_flag"` // VIRTUAL生成列，只读
	CreatedAt      time.Time  `gorm:"type:datetime(3);column:created_at"`
	UpdatedAt      time.Time  `gorm:"type:datetime(3);column:updated_at"`
}

func (UserFollow) TableName() string {
	return "user_follow"
}

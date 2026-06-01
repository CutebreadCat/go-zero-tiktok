package user_follow

// UserFollow 用户关注关系数据库模型
type UserFollow struct {
	FollowerID string `gorm:"primaryKey;type:varchar(64);column:follower_id"`
	UserID     string `gorm:"primaryKey;type:varchar(64);column:user_id"`
	Status     int32  `gorm:"default:0;type:int;column:status"`
}

func (UserFollow) TableName() string {
	return "user_follow"
}

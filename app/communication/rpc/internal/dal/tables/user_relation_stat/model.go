package user_relation_stat

import "time"

// UserRelationStat 用户关系统计数据库模型
// 计数由 communication 服务在 follow/unfollow 操作时原子维护，避免每次实时 COUNT 全表扫描
type UserRelationStat struct {
	UserID         int64     `gorm:"primaryKey;type:bigint;column:user_id"`
	FollowerCount  int64     `gorm:"default:0;type:bigint;column:follower_count"`
	FollowingCount int64     `gorm:"default:0;type:bigint;column:following_count"`
	FriendCount    int64     `gorm:"default:0;type:bigint;column:friend_count"`
	CreatedAt      time.Time `gorm:"type:datetime(3);column:created_at"`
	UpdatedAt      time.Time `gorm:"type:datetime(3);column:updated_at"`
}

func (UserRelationStat) TableName() string {
	return "user_relation_stat"
}

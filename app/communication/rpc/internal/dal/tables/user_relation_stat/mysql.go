package user_relation_stat

import (
	"context"
	"database/sql"
	"errors"

	"go_zero-tiktok/pkg/xerr"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetByUserID 查询用户关系统计
func GetByUserID(ctx context.Context, db *gorm.DB, userID int64) (*UserRelationStat, error) {
	if userID == 0 {
		return nil, xerr.NewInvalidParam("用户ID为空")
	}

	var stat UserRelationStat
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&stat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, xerr.Wrap(err, "get relation stat failed")
	}
	return &stat, nil
}

// GetOrCreate 查询统计行，不存在则初始化（INSERT IGNORE 语义，幂等）
func GetOrCreate(ctx context.Context, db *gorm.DB, userID int64) (*UserRelationStat, error) {
	if userID == 0 {
		return nil, xerr.NewInvalidParam("用户ID为空")
	}

	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&UserRelationStat{UserID: userID}).Error; err != nil {
		return nil, xerr.Wrap(err, "init relation stat failed")
	}
	return GetByUserID(ctx, db, userID)
}

// IncreaseFollowerCount 原子增减粉丝数（行不存在自动初始化）
func IncreaseFollowerCount(ctx context.Context, db *gorm.DB, userID int64, delta int64) error {
	return increaseCount(ctx, db, userID, "follower_count", delta)
}

// IncreaseFollowingCount 原子增减关注数（行不存在自动初始化）
func IncreaseFollowingCount(ctx context.Context, db *gorm.DB, userID int64, delta int64) error {
	return increaseCount(ctx, db, userID, "following_count", delta)
}

// IncreaseFriendCount 原子增减互关数（行不存在自动初始化）
func IncreaseFriendCount(ctx context.Context, db *gorm.DB, userID int64, delta int64) error {
	return increaseCount(ctx, db, userID, "friend_count", delta)
}

// increaseCount 使用 INSERT ... ON DUPLICATE KEY UPDATE 语义，单条语句原子更新，天然幂等。
// 注意：INSERT 时需将目标列的初始值设为 delta（行不存在时直接以 delta 初始化），
// 冲突分支再以 已有值 + delta 累加，保证首次调用也正确应用增量（MySQL/SQLite 行为一致）。
func increaseCount(ctx context.Context, db *gorm.DB, userID int64, column string, delta int64) error {
	if userID == 0 {
		return xerr.NewInvalidParam("用户ID为空")
	}

	stat := &UserRelationStat{UserID: userID}
	switch column {
	case "follower_count":
		stat.FollowerCount = delta
	case "following_count":
		stat.FollowingCount = delta
	case "friend_count":
		stat.FriendCount = delta
	}

	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{column: gorm.Expr(column+" + ?", delta)}),
		}).
		Create(stat).Error; err != nil {
		return xerr.Wrap(err, "increase "+column+" failed")
	}
	return nil
}

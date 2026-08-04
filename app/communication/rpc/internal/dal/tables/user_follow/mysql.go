package user_follow

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go_zero-tiktok/app/communication/rpc/internal/dal/query"
	"go_zero-tiktok/pkg/xerr"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// FollowUser 关注用户（软删除恢复语义 + 幂等）
func FollowUser(ctx context.Context, db *gorm.DB, followerID, userID int64) error {
	if followerID == 0 || userID == 0 {
		return xerr.NewInvalidParam("关注者ID或用户ID为空")
	}
	if followerID == userID {
		return xerr.NewInvalidParam("不能关注自己")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 三段式幂等: 先查活跃关系
		var relation UserFollow
		err := tx.Where("follower_id = ? AND user_id = ? AND deleted_at IS NULL", followerID, userID).First(&relation).Error
		if err == nil {
			return xerr.NewInvalidParam("重复关注")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return xerr.Wrap(err, "query follow relation failed")
		}

		// 软删记录存在则恢复(deleted_at 置 NULL)，active_flag 生成列自动变回 1
		result := tx.Model(&UserFollow{}).
			Where("follower_id = ? AND user_id = ? AND deleted_at IS NOT NULL", followerID, userID).
			Update("deleted_at", nil)
		if result.Error != nil {
			return xerr.Wrap(result.Error, "restore follow relation failed")
		}
		if result.RowsAffected > 0 {
			return nil
		}

		// 无任何记录则新建
		if err := tx.Create(&UserFollow{FollowerID: followerID, UserID: userID}).Error; err != nil {
			// 撞唯一索引 1062: 并发下另一事务已写入，按幂等成功处理
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				return nil
			}
			return xerr.Wrap(err, "create follow relation failed")
		}

		return nil
	})
}

// UnfollowUser 取消关注（软删除: deleted_at 置当前时间，active_flag 生成列自动变 NULL）
func UnfollowUser(ctx context.Context, db *gorm.DB, followerID, userID int64) error {
	if followerID == 0 || userID == 0 {
		return xerr.NewInvalidParam("关注者ID或用户ID为空")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&UserFollow{}).
			Where("follower_id = ? AND user_id = ? AND deleted_at IS NULL", followerID, userID).
			Update("deleted_at", now)
		if result.Error != nil {
			return xerr.Wrap(result.Error, "soft delete follow relation failed")
		}
		if result.RowsAffected == 0 {
			return xerr.NewInvalidParam("未关注该用户")
		}
		return nil
	})
}

func GetFollowingByFollowerID(ctx context.Context, db *gorm.DB, followerID int64, pageNumber, pageSize int32) ([]UserFollow, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&UserFollow{}).
		Where("follower_id = ? AND deleted_at IS NULL", followerID)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "count following list failed")
	}

	var relations []UserFollow
	if err := dbQuery.Scopes(query.Paginate(int(pageNumber), int(pageSize))).
		Find(&relations).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get following list failed")
	}

	return relations, total, nil
}

func GetFansByUserID(ctx context.Context, db *gorm.DB, userID int64, pageNumber, pageSize int32) ([]UserFollow, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&UserFollow{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "count fans list failed")
	}

	var relations []UserFollow
	if err := dbQuery.Scopes(query.Paginate(int(pageNumber), int(pageSize))).
		Find(&relations).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get fans list failed")
	}

	return relations, total, nil
}

// GetFriendByUserID 互关好友 = 我关注了对方 且 对方也关注我（双向活跃关系）
func GetFriendByUserID(ctx context.Context, db *gorm.DB, userID int64, pageNumber, pageSize int32) ([]UserFollow, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&UserFollow{}).
		Where("follower_id = ? AND deleted_at IS NULL", userID).
		Where("EXISTS (SELECT 1 FROM user_follow f2 WHERE f2.follower_id = user_follow.user_id AND f2.user_id = user_follow.follower_id AND f2.deleted_at IS NULL)")

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "count friends list failed")
	}

	var relations []UserFollow
	if err := dbQuery.Scopes(query.Paginate(int(pageNumber), int(pageSize))).
		Find(&relations).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get friends list failed")
	}

	return relations, total, nil
}

// GetActiveRelation 查询活跃关注关系（幂等校验用）
func GetActiveRelation(ctx context.Context, db *gorm.DB, followerID, userID int64) (*UserFollow, error) {
	var relation UserFollow
	err := db.WithContext(ctx).
		Where("follower_id = ? AND user_id = ? AND deleted_at IS NULL", followerID, userID).
		First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, xerr.Wrap(err, "query active follow relation failed")
	}
	return &relation, nil
}

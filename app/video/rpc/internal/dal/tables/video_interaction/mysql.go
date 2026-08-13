package video_interaction

import (
	"context"
	"errors"
	"fmt"

	"go_zero-tiktok/app/video/rpc/internal/dal/query"
	videostattable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_stat"
	"go_zero-tiktok/pkg/xerr"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ActionTypeLike     int32 = 1 // 点赞
	ActionTypeFavorite int32 = 2 // 收藏
)

// AddInteraction 创建交互关系（三段式幂等）
func AddInteraction(ctx context.Context, db *gorm.DB, userID, videoID int64, actionType int32, duplicateMsg string) error {
	if userID == 0 || videoID == 0 {
		return xerr.NewInvalidParam("用户ID或视频ID为空")
	}

	interaction := &VideoInteraction{
		UserID:     userID,
		VideoID:    videoID,
		ActionType: actionType,
	}

	// 三段式幂等：先查
	var existed VideoInteraction
	err := db.WithContext(ctx).Where("user_id = ? AND video_id = ? AND action_type = ?", userID, videoID, actionType).First(&existed).Error
	if err == nil {
		return xerr.NewInvalidParam(duplicateMsg)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.Wrap(err, "check interaction relation failed")
	}

	// 再插
	if err := db.WithContext(ctx).Create(interaction).Error; err != nil {
		// 撞 1062 唯一键后重查，命中则视为幂等成功
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			var existed VideoInteraction
			if err2 := db.WithContext(ctx).Where("user_id = ? AND video_id = ? AND action_type = ?", userID, videoID, actionType).First(&existed).Error; err2 == nil {
				return nil
			}
		}
		return xerr.Wrap(err, "create interaction failed")
	}

	return nil
}

// LikeVideo 点赞视频
func LikeVideo(ctx context.Context, db *gorm.DB, userID, videoID int64) error {
	return AddInteraction(ctx, db, userID, videoID, ActionTypeLike, "重复点赞")
}

// FavoriteVideo 收藏视频
func FavoriteVideo(ctx context.Context, db *gorm.DB, userID, videoID int64) error {
	return AddInteraction(ctx, db, userID, videoID, ActionTypeFavorite, "重复收藏")
}

// RemoveInteraction 删除交互关系
func RemoveInteraction(ctx context.Context, db *gorm.DB, userID, videoID int64, actionType int32, notFoundMsg string) error {
	result := db.WithContext(ctx).
		Where("user_id = ? AND video_id = ? AND action_type = ?", userID, videoID, actionType).
		Delete(&VideoInteraction{})
	if result.Error != nil {
		return xerr.Wrap(result.Error, "cancel interaction failed")
	}

	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam(notFoundMsg)
	}

	return nil
}

// CancelLikeVideo 取消点赞
func CancelLikeVideo(ctx context.Context, db *gorm.DB, userID, videoID int64) error {
	return RemoveInteraction(ctx, db, userID, videoID, ActionTypeLike, "点赞关系不存在")
}

// CancelFavoriteVideo 取消收藏
func CancelFavoriteVideo(ctx context.Context, db *gorm.DB, userID, videoID int64) error {
	return RemoveInteraction(ctx, db, userID, videoID, ActionTypeFavorite, "收藏关系不存在")
}

// GetVideoIDsByUserIDAndAction 分页查询用户某类交互的视频 ID 列表（按交互时间倒序）
func GetVideoIDsByUserIDAndAction(ctx context.Context, db *gorm.DB, userID int64, actionType int32, pageNumber, pageSize int32) ([]int64, int64, error) {
	dbQuery := db.WithContext(ctx).
		Model(&VideoInteraction{}).
		Where("user_id = ? AND action_type = ?", userID, actionType)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "count interactions failed")
	}

	var rows []VideoInteraction
	if err := dbQuery.Order("created_at DESC").Scopes(query.Paginate(int(pageNumber), int(pageSize))).
		Find(&rows).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get interaction video ids failed")
	}

	videoIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		videoIDs = append(videoIDs, row.VideoID)
	}

	return videoIDs, total, nil
}

// GetLikedVideoIDsByUserID 获取用户点赞的视频 ID 列表
func GetLikedVideoIDsByUserID(ctx context.Context, db *gorm.DB, userID int64, pageNumber, pageSize int32) ([]int64, int64, error) {
	return GetVideoIDsByUserIDAndAction(ctx, db, userID, ActionTypeLike, pageNumber, pageSize)
}

// GetFavoritedVideoIDsByUserID 获取用户收藏的视频 ID 列表
func GetFavoritedVideoIDsByUserID(ctx context.Context, db *gorm.DB, userID int64, pageNumber, pageSize int32) ([]int64, int64, error) {
	return GetVideoIDsByUserIDAndAction(ctx, db, userID, ActionTypeFavorite, pageNumber, pageSize)
}

// GetLikeUserIDsByVideoID 获取点赞某视频的全部用户 ID（供 syncer 对比差集）。
func GetLikeUserIDsByVideoID(ctx context.Context, db *gorm.DB, videoID int64) ([]int64, error) {
	var rows []VideoInteraction
	if err := db.WithContext(ctx).
		Where("video_id = ? AND action_type = ?", videoID, ActionTypeLike).
		Find(&rows).Error; err != nil {
		return nil, xerr.Wrap(err, "get like user ids failed")
	}

	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.UserID)
	}
	return ids, nil
}

// BatchAddLikeInteractions 批量插入点赞关系（使用 INSERT IGNORE 忽略已存在记录）。
func BatchAddLikeInteractions(ctx context.Context, db *gorm.DB, videoID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	records := make([]*VideoInteraction, 0, len(userIDs))
	for _, uid := range userIDs {
		records = append(records, &VideoInteraction{
			UserID:     uid,
			VideoID:    videoID,
			ActionType: ActionTypeLike,
		})
	}

	// GORM 的 CreateInBatches + Clauses(ignore) 实现批量 INSERT IGNORE。
	return db.WithContext(ctx).
		Clauses(clause.Insert{Modifier: "IGNORE"}).
		CreateInBatches(records, 500).Error
}

// BatchRemoveLikeInteractions 批量删除点赞关系。
func BatchRemoveLikeInteractions(ctx context.Context, db *gorm.DB, videoID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	return db.WithContext(ctx).
		Where("video_id = ? AND action_type = ? AND user_id IN ?", videoID, ActionTypeLike, userIDs).
		Delete(&VideoInteraction{}).Error
}

// AddLikeInteraction 幂等插入一条点赞关系。
// MySQL 使用 INSERT IGNORE；SQLite 使用 ON CONFLICT DO NOTHING。
// 返回 true 表示实际新增了一条记录；false 表示记录已存在（幂等）。
func AddLikeInteraction(ctx context.Context, db *gorm.DB, userID, videoID int64) (bool, error) {
	if userID == 0 || videoID == 0 {
		return false, xerr.NewInvalidParam("用户ID或视频ID为空")
	}
	record := &VideoInteraction{
		UserID:     userID,
		VideoID:    videoID,
		ActionType: ActionTypeLike,
	}
	result := db.WithContext(ctx).
		Clauses(likeConflictClause(db)).
		Create(record)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// RemoveLikeInteraction 幂等删除一条点赞关系（未命中不报错）。
// 返回 true 表示实际删除了一条记录；false 表示记录不存在（幂等）。
func RemoveLikeInteraction(ctx context.Context, db *gorm.DB, userID, videoID int64) (bool, error) {
	if userID == 0 || videoID == 0 {
		return false, xerr.NewInvalidParam("用户ID或视频ID为空")
	}
	result := db.WithContext(ctx).
		Where("user_id = ? AND video_id = ? AND action_type = ?", userID, videoID, ActionTypeLike).
		Delete(&VideoInteraction{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ApplyLikeEvent 在事务内应用点赞/取消点赞事件。
// action 取值与 interaction.LikeAction 一致："like" / "cancel"。
func ApplyLikeEvent(ctx context.Context, db *gorm.DB, action string, userID, videoID int64) error {
	if userID == 0 || videoID == 0 {
		return xerr.NewInvalidParam("用户ID或视频ID为空")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保证 video_stat 行存在，避免 UpdateVideoLikeCount 因记录缺失失败。
		if err := videostattable.EnsurePopularVideo(ctx, tx, videoID); err != nil {
			return xerr.Wrap(err, "ApplyLikeEvent ensure video_stat")
		}

		switch action {
		case "like":
			added, err := AddLikeInteraction(ctx, tx, userID, videoID)
			if err != nil {
				return err
			}
			if added {
				return videostattable.UpdateVideoLikeCount(ctx, tx, videoID, 1)
			}
			return nil
		case "cancel":
			removed, err := RemoveLikeInteraction(ctx, tx, userID, videoID)
			if err != nil {
				return err
			}
			if removed {
				return videostattable.UpdateVideoLikeCount(ctx, tx, videoID, -1)
			}
			return nil
		default:
			return xerr.NewInvalidParam(fmt.Sprintf("unsupported like action: %s", action))
		}
	})
}

// likeConflictClause 根据数据库方言返回合适的幂等插入子句。
func likeConflictClause(db *gorm.DB) clause.Expression {
	if db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		return clause.OnConflict{DoNothing: true}
	}
	return clause.Insert{Modifier: "IGNORE"}
}

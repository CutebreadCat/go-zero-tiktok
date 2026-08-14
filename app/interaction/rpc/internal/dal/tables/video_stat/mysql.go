package video_stat

import (
	"context"
	"fmt"

	"go_zero-tiktok/app/interaction/rpc/internal/dal/query"
	"go_zero-tiktok/pkg/xerr"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreatePopularVideo(ctx context.Context, db *gorm.DB, videoID int64) error {
	record := &VideoStat{
		VideoID:       videoID,
		VisitCount:    0,
		LikeCount:     0,
		CommentCount:  0,
		FavoriteCount: 0,
	}

	if err := db.WithContext(ctx).Create(record).Error; err != nil {
		return xerr.Wrap(err, "create popular video failed")
	}

	return nil
}

// EnsurePopularVideo 幂等创建 video_stat 行（已存在则忽略）。
// 用于消费端落库前保证 stat 记录存在，避免 UpdateVideoLikeCount 因 RowsAffected=0 失败。
func EnsurePopularVideo(ctx context.Context, db *gorm.DB, videoID int64) error {
	record := &VideoStat{
		VideoID:       videoID,
		VisitCount:    0,
		LikeCount:     0,
		CommentCount:  0,
		FavoriteCount: 0,
	}
	return db.WithContext(ctx).
		Clauses(statConflictClause(db)).
		Create(record).Error
}

// statConflictClause 根据数据库方言返回合适的幂等插入子句。
func statConflictClause(db *gorm.DB) clause.Expression {
	if db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		return clause.OnConflict{DoNothing: true}
	}
	return clause.Insert{Modifier: "IGNORE"}
}

func IncreaseVideoVisitCount(ctx context.Context, db *gorm.DB, videoID int64, delta int64) error {
	if delta <= 0 {
		delta = 1
	}

	result := db.WithContext(ctx).
		Model(&VideoStat{}).
		Where("video_id = ?", videoID).
		Update("visit_count", gorm.Expr("visit_count + ?", delta))
	if result.Error != nil {
		return xerr.Wrap(result.Error, "increase video visit count failed")
	}

	return nil
}

func UpdateVideoLikeCount(ctx context.Context, db *gorm.DB, videoID int64, delta int64) error {
	result := db.WithContext(ctx).
		Model(&VideoStat{}).
		Where("video_id = ?", videoID).
		Update("like_count", gorm.Expr("CASE WHEN like_count + ? < 0 THEN 0 ELSE like_count + ? END", delta, delta))
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update video like count failed")
	}

	if result.RowsAffected == 0 {
		return xerr.Wrap(fmt.Errorf("video %d not found", videoID), "update video like count failed")
	}

	return nil
}

func UpdateVideoFavoriteCount(ctx context.Context, db *gorm.DB, videoID int64, delta int64) error {
	result := db.WithContext(ctx).
		Model(&VideoStat{}).
		Where("video_id = ?", videoID).
		Update("favorite_count", gorm.Expr("CASE WHEN favorite_count + ? < 0 THEN 0 ELSE favorite_count + ? END", delta, delta))
	if result.Error != nil {
		return xerr.Wrap(result.Error, "update video favorite count failed")
	}

	if result.RowsAffected == 0 {
		return xerr.Wrap(fmt.Errorf("video %d not found", videoID), "update video favorite count failed")
	}

	return nil
}

func SetLikeCount(ctx context.Context, db *gorm.DB, videoID int64, count int64) error {
	result := db.WithContext(ctx).
		Model(&VideoStat{}).
		Where("video_id = ?", videoID).
		Update("like_count", count)
	if result.Error != nil {
		return xerr.Wrap(result.Error, "set video like count failed")
	}
	if result.RowsAffected == 0 {
		return xerr.Wrap(fmt.Errorf("video %d not found", videoID), "set video like count failed")
	}
	return nil
}

func GetLikeCounts(ctx context.Context, db *gorm.DB, videoIDs []int64) (map[int64]int64, error) {
	if len(videoIDs) == 0 {
		return map[int64]int64{}, nil
	}

	var rows []VideoStat
	if err := db.WithContext(ctx).Where("video_id IN ?", videoIDs).Find(&rows).Error; err != nil {
		return nil, xerr.Wrap(err, "get like counts failed")
	}

	result := make(map[int64]int64, len(rows))
	for _, row := range rows {
		result[row.VideoID] = row.LikeCount
	}

	// 未找到的视频默认返回 0，避免调用方需要二次判断。
	for _, id := range videoIDs {
		if _, ok := result[id]; !ok {
			result[id] = 0
		}
	}

	return result, nil
}

func GetFavoriteCounts(ctx context.Context, db *gorm.DB, videoIDs []int64) (map[int64]int64, error) {
	if len(videoIDs) == 0 {
		return map[int64]int64{}, nil
	}

	var rows []VideoStat
	if err := db.WithContext(ctx).Where("video_id IN ?", videoIDs).Find(&rows).Error; err != nil {
		return nil, xerr.Wrap(err, "get favorite counts failed")
	}

	result := make(map[int64]int64, len(rows))
	for _, row := range rows {
		result[row.VideoID] = row.FavoriteCount
	}

	for _, id := range videoIDs {
		if _, ok := result[id]; !ok {
			result[id] = 0
		}
	}

	return result, nil
}

func GetPopularVideoIDsByVisitCount(ctx context.Context, db *gorm.DB, pageNum, pageSize int32) ([]VideoStat, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&VideoStat{})

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get popular video count failed")
	}

	var rows []VideoStat
	if err := dbQuery.Order("visit_count DESC").Scopes(query.Paginate(int(pageNum), int(pageSize))).
		Find(&rows).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get popular videos failed")
	}

	return rows, total, nil
}

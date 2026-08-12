package video_favoriter

import (
	"context"
	"errors"

	"go_zero-tiktok/app/video/rpc/internal/dal/query"
	"go_zero-tiktok/pkg/xerr"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func FavoriteVideo(ctx context.Context, db *gorm.DB, userID, videoID int64) error {
	if userID == 0 || videoID == 0 {
		return xerr.NewInvalidParam("用户ID或视频ID为空")
	}

	fav := &VideoFavoriter{
		UserID:  userID,
		VideoID: videoID,
	}

	// 三段式幂等：先查
	var existed VideoFavoriter
	err := db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).First(&existed).Error
	if err == nil {
		return xerr.NewInvalidParam("重复收藏")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.Wrap(err, "check favorite relation failed")
	}

	// 再插
	if err := db.WithContext(ctx).Create(fav).Error; err != nil {
		// 撞 1062 唯一键后重查，命中则视为幂等成功
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			var existed VideoFavoriter
			if err2 := db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).First(&existed).Error; err2 == nil {
				return nil
			}
		}
		return xerr.Wrap(err, "favorite video failed")
	}

	return nil
}

func CancelFavoriteVideo(ctx context.Context, db *gorm.DB, userID, videoID int64) error {
	result := db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&VideoFavoriter{})
	if result.Error != nil {
		return xerr.Wrap(result.Error, "cancel favorite video failed")
	}

	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("收藏关系不存在")
	}

	return nil
}

func GetFavoritedVideoIDsByUserID(ctx context.Context, db *gorm.DB, userID int64, pageNumber, pageSize int32) ([]int64, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&VideoFavoriter{}).Where("user_id = ?", userID)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get favorited video ids count failed")
	}

	var favRows []VideoFavoriter
	if err := dbQuery.Order("created_at DESC").Scopes(query.Paginate(int(pageNumber), int(pageSize))).
		Find(&favRows).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get favorited video ids failed")
	}

	videoIDs := make([]int64, 0, len(favRows))
	for _, row := range favRows {
		videoIDs = append(videoIDs, row.VideoID)
	}

	return videoIDs, total, nil
}

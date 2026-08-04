package video_liker

import (
	"context"
	"errors"

	"go_zero-tiktok/app/video/rpc/internal/dal/query"
	"go_zero-tiktok/pkg/xerr"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func LikeVideo(ctx context.Context, db *gorm.DB, userID, videoID int64) error {
	if userID == 0 || videoID == 0 {
		return xerr.NewInvalidParam("用户ID或视频ID为空")
	}

	like := &VideoLiker{
		UserID:  userID,
		VideoID: videoID,
	}

	// 三段式幂等：先查
	var existed VideoLiker
	err := db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).First(&existed).Error
	if err == nil {
		return xerr.NewInvalidParam("重复点赞")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.Wrap(err, "check like relation failed")
	}

	// 再插
	if err := db.WithContext(ctx).Create(like).Error; err != nil {
		// 撞 1062 唯一键后重查，命中则视为幂等成功
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			var existed VideoLiker
			if err2 := db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).First(&existed).Error; err2 == nil {
				return nil
			}
		}
		return xerr.Wrap(err, "like video failed")
	}

	return nil
}

func CancelLikeVideo(ctx context.Context, db *gorm.DB, userID, videoID int64) error {
	result := db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&VideoLiker{})
	if result.Error != nil {
		return xerr.Wrap(result.Error, "cancel like video failed")
	}

	if result.RowsAffected == 0 {
		return xerr.NewInvalidParam("点赞关系不存在")
	}

	return nil
}

func GetLikedVideoIDsByUserID(ctx context.Context, db *gorm.DB, userID int64, pageNumber, pageSize int32) ([]int64, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&VideoLiker{}).Where("user_id = ?", userID)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get liked video ids count failed")
	}

	var likerRows []VideoLiker
	if err := dbQuery.Scopes(query.Paginate(int(pageNumber), int(pageSize))).
		Find(&likerRows).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get liked video ids failed")
	}

	videoIDs := make([]int64, 0, len(likerRows))
	for _, row := range likerRows {
		videoIDs = append(videoIDs, row.VideoID)
	}

	return videoIDs, total, nil
}

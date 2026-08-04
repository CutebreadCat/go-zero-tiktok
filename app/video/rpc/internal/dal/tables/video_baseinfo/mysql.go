package video_baseinfo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go_zero-tiktok/app/video/rpc/internal/dal/query"
	"go_zero-tiktok/pkg/xerr"

	myutils "go_zero-tiktok/pkg/utils"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func CreateVideo(ctx context.Context, db *gorm.DB, video *VideoBaseinfo) error {
	// 三段式幂等：先查
	var existed VideoBaseinfo
	err := db.WithContext(ctx).Where("video_id = ?", video.VideoID).First(&existed).Error
	if err == nil {
		// 幂等：同一视频已发布，直接返回成功
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.Wrap(err, "check video exist failed")
	}

	// 再插
	if err := db.WithContext(ctx).Create(video).Error; err != nil {
		// 撞 1062 唯一键后重查，命中则视为幂等成功
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			var exist VideoBaseinfo
			if err2 := db.WithContext(ctx).Where("video_id = ?", video.VideoID).First(&exist).Error; err2 == nil {
				return nil
			}
		}
		return xerr.Wrap(err, "create video failed")
	}

	return nil
}

func SearchVideosByKeyword(ctx context.Context, db *gorm.DB, keyword string, pageNum, pageSize int32) ([]VideoBaseinfo, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&VideoBaseinfo{})
	if keyword != "" {
		dbQuery = dbQuery.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "search videos count failed")
	}

	var videos []VideoBaseinfo
	if err := dbQuery.Order("created_at DESC").Scopes(query.Paginate(int(pageNum), int(pageSize))).
		Find(&videos).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "search videos failed")
	}

	return videos, total, nil
}

func GetVideosByIDs(ctx context.Context, db *gorm.DB, videoIDs []int64) ([]VideoBaseinfo, error) {
	if len(videoIDs) == 0 {
		return []VideoBaseinfo{}, nil
	}
	quotedIDs := make([]string, len(videoIDs))
	for i, id := range videoIDs {
		quotedIDs[i] = strconv.FormatInt(id, 10)
	}
	idsForOrder := strings.Join(quotedIDs, ",")

	var videos []VideoBaseinfo
	if err := db.WithContext(ctx).
		Where("video_id IN ?", videoIDs).
		Order(fmt.Sprintf("FIELD(video_id, %s)", idsForOrder)).
		Find(&videos).Error; err != nil {
		return nil, xerr.Wrap(err, "get videos by ids failed")
	}

	return videos, nil
}

func GetVideosByAuthorID(ctx context.Context, db *gorm.DB, authorID int64, pageNum, pageSize int32) ([]VideoBaseinfo, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&VideoBaseinfo{}).Where("author_id = ?", authorID)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get videos by author count failed")
	}

	var videos []VideoBaseinfo
	if err := dbQuery.Order("created_at DESC").Scopes(query.Paginate(int(pageNum), int(pageSize))).
		Find(&videos).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get videos by author failed")
	}

	return videos, total, nil
}

func GetVideoByLastTime(ctx context.Context, db *gorm.DB, lastTime string, pageNum, pageSize int32) ([]VideoBaseinfo, int64, error) {
	dbQuery := db.WithContext(ctx).Model(&VideoBaseinfo{})
	lastTime = strings.TrimSpace(lastTime)
	if lastTime != "" {
		RealTime, err := myutils.StrToTime(lastTime, "")
		if err != nil {
			return nil, 0, xerr.Wrap(err, "parse last time failed")
		}
		dbQuery = dbQuery.Where("created_at < ?", RealTime)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get videos by last time count failed")
	}

	var videos []VideoBaseinfo
	if err := dbQuery.Scopes(query.Paginate(int(pageNum), int(pageSize))).
		Find(&videos).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get videos by last time failed")
	}
	return videos, total, nil
}

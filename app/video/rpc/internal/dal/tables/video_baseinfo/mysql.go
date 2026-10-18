package video_baseinfo

import (
	"context"
	"errors"
	"sort"
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

// GetVideoByAuthorAndIdempotencyKey 按 (author_id, idempotency_key) 查询已发布视频。
// 未命中返回 (nil, nil)，由调用方按"不存在"处理，避免泄漏 gorm.ErrRecordNotFound。
func GetVideoByAuthorAndIdempotencyKey(ctx context.Context, db *gorm.DB, authorID int64, idempotencyKey string) (*VideoBaseinfo, error) {
	if idempotencyKey == "" {
		return nil, nil
	}

	var video VideoBaseinfo
	err := db.WithContext(ctx).
		Select("video_id, author_id, created_at").
		Where("author_id = ? AND idempotency_key = ?", authorID, idempotencyKey).
		First(&video).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, xerr.Wrap(err, "get video by idempotency key failed")
	}

	return &video, nil
}

// CreateVideoIdempotent 带幂等键创建视频，返回最终生效的 video_id。
// 同一 (author_id, idempotency_key) 重复提交（含并发）时复用已存在记录，不重复插入；
// 幂等键为空时退化为 CreateVideo 的原行为。
func CreateVideoIdempotent(ctx context.Context, db *gorm.DB, video *VideoBaseinfo) (int64, error) {
	if video.IdempotencyKey == nil || *video.IdempotencyKey == "" {
		if err := CreateVideo(ctx, db, video); err != nil {
			return 0, err
		}
		return video.VideoID, nil
	}

	// 三段式幂等：先查
	existed, err := GetVideoByAuthorAndIdempotencyKey(ctx, db, video.AuthorID, *video.IdempotencyKey)
	if err != nil {
		return 0, err
	}
	if existed != nil {
		return existed.VideoID, nil
	}

	// 再插
	if err := db.WithContext(ctx).Create(video).Error; err != nil {
		// 处理重复：并发请求撞 uk_author_idempotency 后回查，返回先写入的那条
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			if existed, err2 := GetVideoByAuthorAndIdempotencyKey(ctx, db, video.AuthorID, *video.IdempotencyKey); err2 == nil && existed != nil {
				return existed.VideoID, nil
			}
			// 撞 video_id 主键（同 ID 重放）同样视为幂等成功
			var sameID VideoBaseinfo
			if err2 := db.WithContext(ctx).Where("video_id = ?", video.VideoID).First(&sameID).Error; err2 == nil {
				return sameID.VideoID, nil
			}
		}
		return 0, xerr.Wrap(err, "create video failed")
	}

	return video.VideoID, nil
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

// feedQueryColumns 水合查询限定字段：排除 idempotency_key 等列表展示无关列，降低行宽。
// 注意需与 VideoToResponse 用到的字段保持一致。
const feedQueryColumns = "video_id, author_id, video_object_key, cover_object_key, title, description, created_at, updated_at, deleted_at"

func GetVideosByIDs(ctx context.Context, db *gorm.DB, videoIDs []int64) ([]VideoBaseinfo, error) {
	if len(videoIDs) == 0 {
		return []VideoBaseinfo{}, nil
	}

	var videos []VideoBaseinfo
	if err := db.WithContext(ctx).
		Select(feedQueryColumns).
		Where("video_id IN ?", videoIDs).
		Where("deleted_at IS NULL").
		Find(&videos).Error; err != nil {
		return nil, xerr.Wrap(err, "get videos by ids failed")
	}

	// 按传入 ID 顺序排序（避免依赖 MySQL 专有 FIELD() 函数，保证跨数据库一致）
	order := make(map[int64]int, len(videoIDs))
	for i, id := range videoIDs {
		order[id] = i
	}
	sort.SliceStable(videos, func(i, j int) bool {
		return order[videos[i].VideoID] < order[videos[j].VideoID]
	})

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
	// 兜底查询与 Redis 索引路径保持一致的软删过滤，避免已删视频从兜底路径漏出。
	dbQuery := db.WithContext(ctx).Model(&VideoBaseinfo{}).Where("deleted_at IS NULL")
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

// GetVideosByCursor 复合游标分页：按 (created_at, video_id) < (publishedAt, videoID) 倒序取 limit 条。
// publishedAt=0 且 videoID=0 表示首页，不附加游标条件。
func GetVideosByCursor(ctx context.Context, db *gorm.DB, publishedAt, videoID int64, limit int32) ([]VideoBaseinfo, error) {
	dbQuery := db.WithContext(ctx).
		Model(&VideoBaseinfo{}).
		Select(feedQueryColumns).
		Where("deleted_at IS NULL")

	if publishedAt > 0 || videoID > 0 {
		dbQuery = dbQuery.Where("(created_at, video_id) < (?, ?)", publishedAt, videoID)
	}

	var videos []VideoBaseinfo
	if err := dbQuery.
		Order("created_at DESC, video_id DESC").
		Limit(int(limit)).
		Find(&videos).Error; err != nil {
		return nil, xerr.Wrap(err, "get videos by cursor failed")
	}
	return videos, nil
}

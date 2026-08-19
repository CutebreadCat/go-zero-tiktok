package reposity

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	videobasetable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_baseinfo"
	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/storage/aliyun"
	myutils "go_zero-tiktok/pkg/utils"

	pkgerrors "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/syncx"
	"gorm.io/gorm"
)

// videoDetailCacheExpire 视频详情本地缓存有效期。
// 视频详情几乎不变（标题/描述编辑是低频操作），短 TTL 保证内存有界 + 变更后快速自愈。
const videoDetailCacheExpire = time.Minute

// videoDetailCacheLimit 本地缓存条目上限，超出后按 LRU 淘汰最久未使用的 key。
const videoDetailCacheLimit = 10000

type VideoBaseinfoRepo struct {
	db    *gorm.DB
	cache *collection.Cache
	sf    syncx.SingleFlight
}

func NewVideoBaseinfoRepo(db *gorm.DB) (*VideoBaseinfoRepo, error) {
	cache, err := collection.NewCache(videoDetailCacheExpire, collection.WithLimit(videoDetailCacheLimit))
	if err != nil {
		return nil, err
	}
	return &VideoBaseinfoRepo{db: db, cache: cache, sf: syncx.NewSingleFlight()}, nil
}

func (r *VideoBaseinfoRepo) CreateVideo(ctx context.Context, video *videobasetable.VideoBaseinfo) error {
	if err := videobasetable.CreateVideo(ctx, r.db, video); err != nil {
		return pkgerrors.WithMessage(err, "VideoBaseinfoRepo.CreateVideo")
	}
	return nil
}

func (r *VideoBaseinfoRepo) CreateVideoFromParams(ctx context.Context, videoID, authorID int64, videoObjectKey, coverObjectKey, title, description string) error {
	video := &videobasetable.VideoBaseinfo{
		VideoID:        videoID,
		AuthorID:       authorID,
		VideoObjectKey: videoObjectKey,
		CoverObjectKey: coverObjectKey,
		Title:          title,
		Description:    description,
	}
	if err := videobasetable.CreateVideo(ctx, r.db, video); err != nil {
		return pkgerrors.WithMessage(err, "VideoBaseinfoRepo.CreateVideoFromParams")
	}
	return nil
}

func (r *VideoBaseinfoRepo) SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := videobasetable.SearchVideosByKeyword(ctx, r.db, keyword, pageNum, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.SearchVideosByKeyword")
	}
	return r.VideosToResponse(videos), total, nil
}

func (r *VideoBaseinfoRepo) GetVideosByIDs(ctx context.Context, videoIDs []int64) ([]types.VideoBaseinfo, error) {
	if len(videoIDs) == 0 {
		return []types.VideoBaseinfo{}, nil
	}

	// ① 批量读缓存（纯内存 O(1) 每条），收集 miss 的 id（去重）
	videoMap := make(map[int64]videobasetable.VideoBaseinfo, len(videoIDs))
	missIDs := make([]int64, 0, len(videoIDs))
	seen := make(map[int64]struct{}, len(videoIDs))
	for _, id := range videoIDs {
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		if v, ok := r.getFromCache(id); ok {
			if v != nil {
				videoMap[id] = *v
			}
			continue
		}
		missIDs = append(missIDs, id)
	}

	// ② miss 的批量兜底：singleflight 合并并发 + 一条 IN SQL 查库并写回缓存
	if len(missIDs) > 0 {
		if err := r.loadBatch(ctx, missIDs); err != nil {
			return nil, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.GetVideosByIDs")
		}
		for _, id := range missIDs {
			if v, ok := r.getFromCache(id); ok && v != nil {
				videoMap[id] = *v
			}
		}
	}

	// ③ 按传入顺序组装
	result := make([]videobasetable.VideoBaseinfo, 0, len(videoIDs))
	for _, id := range videoIDs {
		if v, ok := videoMap[id]; ok {
			result = append(result, v)
		}
	}
	return r.VideosToResponse(result), nil
}

// videoDetailCacheKey 生成单条视频详情缓存的 key。
func videoDetailCacheKey(videoID int64) string {
	return fmt.Sprintf("video_detail:%d", videoID)
}

// getFromCache 读单条视频缓存。
// 返回 ok=true 且 val=nil 表示"负缓存命中"（已知视频不存在）；ok=false 表示缓存 miss。
func (r *VideoBaseinfoRepo) getFromCache(videoID int64) (*videobasetable.VideoBaseinfo, bool) {
	val, ok := r.cache.Get(videoDetailCacheKey(videoID))
	if !ok {
		return nil, false
	}
	if val == nil {
		return nil, true
	}
	return val.(*videobasetable.VideoBaseinfo), true
}

func (r *VideoBaseinfoRepo) setCache(videoID int64, video *videobasetable.VideoBaseinfo) {
	r.cache.Set(videoDetailCacheKey(videoID), video)
}

// loadBatch 批量加载 miss 的 id 到缓存。
// singleflight 按"排序后的 id 集合"合并相同集合的并发请求（集合相同、顺序不同也能合并）；
// 真正执行者 double-check 过滤掉等锁期间已被其他请求写回的 id，再一条 IN SQL 批量查库并逐个写回；
// 所有调用方（执行者 + 等待者）随后统一从缓存读结果，天然 double-check。
func (r *VideoBaseinfoRepo) loadBatch(ctx context.Context, ids []int64) error {
	_, err := r.sf.Do(batchSFKey(ids), func() (any, error) {
		// double-check：可能已有并发请求在等锁期间把部分 id 写回缓存
		remain := make([]int64, 0, len(ids))
		for _, id := range ids {
			if _, ok := r.getFromCache(id); !ok {
				remain = append(remain, id)
			}
		}
		if len(remain) == 0 {
			return nil, nil
		}

		rows, err := videobasetable.GetVideosByIDs(ctx, r.db, remain)
		if err != nil {
			return nil, err
		}
		loaded := make(map[int64]struct{}, len(rows))
		for i := range rows {
			r.setCache(rows[i].VideoID, &rows[i])
			loaded[rows[i].VideoID] = struct{}{}
		}
		// 负缓存：查询结果里没有的 id 视为不存在，缓存 nil 避免反复打库
		for _, id := range remain {
			if _, ok := loaded[id]; !ok {
				r.setCache(id, nil)
			}
		}
		return nil, nil
	})
	return err
}

// batchSFKey 生成 singleflight 合并 key：id 排序后拼接，集合相同但顺序不同的请求可合并。
func batchSFKey(ids []int64) string {
	sorted := append([]int64(nil), ids...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	strs := make([]string, len(sorted))
	for i, id := range sorted {
		strs[i] = strconv.FormatInt(id, 10)
	}
	return "videos:" + strings.Join(strs, ",")
}

// InvalidateVideoCache 主动失效指定视频的本地缓存，供编辑/软删视频后调用。
func (r *VideoBaseinfoRepo) InvalidateVideoCache(videoIDs ...int64) {
	for _, id := range videoIDs {
		r.cache.Del(videoDetailCacheKey(id))
	}
}

func (r *VideoBaseinfoRepo) GetVideosByAuthorID(ctx context.Context, authorID int64, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := videobasetable.GetVideosByAuthorID(ctx, r.db, authorID, pageNum, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.GetVideosByAuthorID")
	}
	return r.VideosToResponse(videos), total, nil
}

func (r *VideoBaseinfoRepo) GetVideoByLastTime(ctx context.Context, lastTime string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	videos, total, err := videobasetable.GetVideoByLastTime(ctx, r.db, lastTime, pageNum, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.GetVideoByLastTime")
	}
	return r.VideosToResponse(videos), total, nil
}

func (r *VideoBaseinfoRepo) GetVideosByCursor(ctx context.Context, publishedAt, videoID int64, limit int32) ([]types.VideoBaseinfo, error) {
	videos, err := videobasetable.GetVideosByCursor(ctx, r.db, publishedAt, videoID, limit)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "VideoBaseinfoRepo.GetVideosByCursor")
	}
	return r.VideosToResponse(videos), nil
}

func (r *VideoBaseinfoRepo) VideoToResponse(video *videobasetable.VideoBaseinfo) types.VideoBaseinfo {
	videoURL := aliyun.BuildURL(video.VideoObjectKey)
	coverURL := ""
	if video.CoverObjectKey != "" {
		coverURL = aliyun.BuildURL(video.CoverObjectKey)
	}

	return types.VideoBaseinfo{
		VideoID:     video.VideoID,
		AuthorID:    video.AuthorID,
		VideoURL:    videoURL,
		CoverURL:    coverURL,
		Title:       video.Title,
		Description: video.Description,
		CreatedAt:   myutils.TimeToStr(video.CreatedAt, ""),
		UpdatedAt:   myutils.TimeToStr(video.UpdatedAt, ""),
		DeletedAt:   myutils.NullTimeToStr(myutils.TimeToNullTime(video.DeletedAt), ""),
	}
}

func (r *VideoBaseinfoRepo) VideosToResponse(videos []videobasetable.VideoBaseinfo) []types.VideoBaseinfo {
	result := make([]types.VideoBaseinfo, 0, len(videos))
	for _, v := range videos {
		result = append(result, r.VideoToResponse(&v))
	}
	return result
}

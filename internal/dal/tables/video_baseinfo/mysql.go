package video_baseinfo

import (
	"context"

	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func CreateVideo(ctx context.Context, db *gorm.DB, video *types.VideoBaseinfo) error {
	logger := logx.WithContext(ctx)

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&types.VideoBaseinfo{}).Create(video).Error; err != nil {
			return xerr.New(1002, "创建视频失败")
		}

		popular := &types.VideoPopular{
			VideoID:      video.VideoID,
			VisitCount:   0,
			LikeCount:    0,
			CommentCount: 0,
		}
		if err := tx.WithContext(ctx).Model(&types.VideoPopular{}).Create(popular).Error; err != nil {
			return xerr.New(1002, "创建热门视频记录失败")
		}

		return nil
	})
	if err != nil {
		logger.Errorf("create video transaction failed: %v", err)
		return xerr.New(1002, "创建视频事务失败")
	}
	return nil
}

func SearchVideosByKeyword(ctx context.Context, db *gorm.DB, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.VideoBaseinfo{})
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("search videos count failed: %v", err)
		return nil, 0, xerr.New(1002, "搜索视频总数失败")
	}

	var videos []types.VideoBaseinfo
	offset := (pageNum - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&videos).Error; err != nil {
		logger.Errorf("search videos failed: %v", err)
		return nil, 0, xerr.New(1002, "搜索视频失败")
	}

	return videos, total, nil
}

func GetVideosByIDs(ctx context.Context, db *gorm.DB, videoIDs []string) ([]types.VideoBaseinfo, error) {
	logger := logx.WithContext(ctx)
	if len(videoIDs) == 0 {
		return []types.VideoBaseinfo{}, nil
	}
	var videos []types.VideoBaseinfo
	// 1. 【新增】手动给每个 ID 加上单引号
	// 比如把 v123 变成 'v123'
	quotedIDs := make([]string, len(videoIDs))
	for i, id := range videoIDs {
		quotedIDs[i] = fmt.Sprintf("'%s'", id)
	}
	// 用逗号连接：'v123','v456','v789'
	idsForOrder := strings.Join(quotedIDs, ",")

	// 2. 执行查询
	// 直接把拼好带引号的字符串塞进 SQL
	if err := db.WithContext(ctx).
		Where("video_id IN ?", videoIDs).
		Order(fmt.Sprintf("FIELD(video_id, %s)", idsForOrder)). // ✅ 修正：使用处理后的 idsForOrder
		Find(&videos).Error; err != nil {
		logger.Errorf("get videos by ids failed: %v", err)
		return nil, xerr.New(1002, "获取视频失败")
	}

	return videos, nil
}

func GetVideosByAuthorID(ctx context.Context, db *gorm.DB, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.VideoBaseinfo{}).Where("author_id = ?", authorID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get videos by author count failed: %v", err)
		return nil, 0, xerr.New(1002, "获取作者视频总数失败")
	}

	var videos []types.VideoBaseinfo
	offset := (pageNum - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&videos).Error; err != nil {
		logger.Errorf("get videos by author failed: %v", err)
		return nil, 0, xerr.New(1002, "获取作者视频失败")
	}

	return videos, total, nil
}

func GetVideosByVisitCount(ctx context.Context, db *gorm.DB, pageNum, pageSize int32, videoIDs []string) ([]types.VideoBaseinfo, int64, error) {
	logger := logx.WithContext(ctx)

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&types.VideoBaseinfo{})
	if len(videoIDs) != 0 {
		query = query.Where("video_id IN ?", videoIDs)
	} else {
		query = query.Order("visit_count DESC")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("get videos by visit count count failed: %v", err)
		return nil, 0, xerr.New(1002, "获取热门视频总数失败")
	}

	var videos []types.VideoBaseinfo
	offset := (pageNum - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&videos).Error; err != nil {
		logger.Errorf("get videos by visit count failed: %v", err)
		return nil, 0, xerr.New(1002, "获取热门视频失败")
	}
	return videos, total, nil
}

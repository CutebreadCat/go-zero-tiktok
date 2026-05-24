package video_baseinfo

import (
	"context"
	"fmt"
	"strings"

	"go_zero-tiktok/internal/shared/xerr"

	"gorm.io/gorm"
)

func CreateVideo(ctx context.Context, db *gorm.DB, video *VideoBaseinfo) error {
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&VideoBaseinfo{}).Create(video).Error; err != nil {
			return xerr.Wrap(err, "create video record failed")
		}

		popular := &VideoPopular{
			VideoID:      video.VideoID,
			VisitCount:   0,
			LikeCount:    0,
			CommentCount: 0,
		}
		if err := tx.WithContext(ctx).Model(&VideoPopular{}).Create(popular).Error; err != nil {
			return xerr.Wrap(err, "create popular video record failed")
		}

		return nil
	})
	if err != nil {
		return xerr.Wrap(err, "create video transaction failed")
	}
	return nil
}

func SearchVideosByKeyword(ctx context.Context, db *gorm.DB, keyword string, pageNum, pageSize int32) ([]VideoBaseinfo, int64, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&VideoBaseinfo{})
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "search videos count failed")
	}

	var videos []VideoBaseinfo
	offset := (pageNum - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&videos).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "search videos failed")
	}

	return videos, total, nil
}

func GetVideosByIDs(ctx context.Context, db *gorm.DB, videoIDs []string) ([]VideoBaseinfo, error) {
	if len(videoIDs) == 0 {
		return []VideoBaseinfo{}, nil
	}
	quotedIDs := make([]string, len(videoIDs))
	for i, id := range videoIDs {
		quotedIDs[i] = fmt.Sprintf("'%s'", id)
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

func GetVideosByAuthorID(ctx context.Context, db *gorm.DB, authorID string, pageNum, pageSize int32) ([]VideoBaseinfo, int64, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&VideoBaseinfo{}).Where("author_id = ?", authorID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get videos by author count failed")
	}

	var videos []VideoBaseinfo
	offset := (pageNum - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&videos).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get videos by author failed")
	}

	return videos, total, nil
}

func GetVideosByVisitCount(ctx context.Context, db *gorm.DB, pageNum, pageSize int32, videoIDs []string) ([]VideoBaseinfo, int64, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.WithContext(ctx).Model(&VideoBaseinfo{})
	if len(videoIDs) != 0 {
		query = query.Where("video_id IN ?", videoIDs)
	} else {
		query = query.Order("visit_count DESC")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get videos by visit count count failed")
	}

	var videos []VideoBaseinfo
	offset := (pageNum - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&videos).Error; err != nil {
		return nil, 0, xerr.Wrap(err, "get videos by visit count failed")
	}
	return videos, total, nil
}

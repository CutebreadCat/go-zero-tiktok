// Package video_rank 提供热榜重建所需的候选读取模型。
//
// 不使用联表：以 video_baseinfo 为驱动表按索引游标翻页（窗口模式走
// (created_at, video_id) 复合索引），再用主键 IN 批量补齐 video_stat 计数，
// 避免大表 JOIN 带来的临时表/回表放大。独立成包，避免把跨表读取塞进单表 DAL。
package video_rank

import (
	"context"
	"time"

	"go_zero-tiktok/pkg/xerr"

	"gorm.io/gorm"
)

const (
	// defaultCandidateLimit 未指定 limit 时的单批扫描行数。
	defaultCandidateLimit int32 = 500
	// statQueryChunk 单次 IN 查询携带的主键数量，避免 SQL 参数过多。
	statQueryChunk = 500
)

// CandidateRow 热榜重建的候选行：视频 ID、四类计数与发布时间。
type CandidateRow struct {
	VideoID       int64     `gorm:"column:video_id"`
	VisitCount    int64     `gorm:"column:visit_count"`
	LikeCount     int64     `gorm:"column:like_count"`
	CommentCount  int64     `gorm:"column:comment_count"`
	FavoriteCount int64     `gorm:"column:favorite_count"`
	PublishAt     time.Time `gorm:"column:publish_at"`
}

// candidateVideo 驱动表扫描出的一行：视频 ID 与发布时间。
type candidateVideo struct {
	VideoID   int64     `gorm:"column:video_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// statRow video_stat 的计数快照。
type statRow struct {
	VideoID       int64 `gorm:"column:video_id"`
	VisitCount    int64 `gorm:"column:visit_count"`
	LikeCount     int64 `gorm:"column:like_count"`
	CommentCount  int64 `gorm:"column:comment_count"`
	FavoriteCount int64 `gorm:"column:favorite_count"`
}

// ListCandidates 取一批候选视频。
//
//	since 为零值：全量模式，按 video_id 游标翻页；
//	since 非零值：窗口模式，按 (created_at, video_id) 复合游标翻页，
//	              且 created_at >= since 的过滤发生在索引区间扫描阶段。
//
// 调用方每次用上一批最后一行的 (PublishAt, VideoID) 作为下一次游标；
// 窗口模式首页传 (since, 0)。
//
// 缺少 video_stat 记录的脏视频会被跳过；若整批都被跳过，会继续向后扫描，
// 保证返回空切片等价于"没有更多候选"，不会让调用方提前结束重建。
func ListCandidates(ctx context.Context, db *gorm.DB, since, cursorAt time.Time, cursorID int64, limit int32) ([]CandidateRow, error) {
	if limit <= 0 {
		limit = defaultCandidateLimit
	}

	rows := make([]CandidateRow, 0, limit)
	scanAt, scanID := cursorAt, cursorID

	for int32(len(rows)) < limit {
		videos, err := listCandidateVideos(ctx, db, since, scanAt, scanID, limit-int32(len(rows)))
		if err != nil {
			return nil, err
		}
		if len(videos) == 0 {
			break
		}

		last := videos[len(videos)-1]
		scanAt, scanID = last.CreatedAt, last.VideoID

		videoIDs := make([]int64, 0, len(videos))
		for _, video := range videos {
			videoIDs = append(videoIDs, video.VideoID)
		}
		stats, err := loadStatRows(ctx, db, videoIDs)
		if err != nil {
			return nil, err
		}

		for _, video := range videos {
			stat, ok := stats[video.VideoID]
			if !ok {
				// 统计行缺失：与内联时被过滤的行为保持一致，跳过该视频。
				continue
			}
			rows = append(rows, CandidateRow{
				VideoID:       video.VideoID,
				VisitCount:    stat.VisitCount,
				LikeCount:     stat.LikeCount,
				CommentCount:  stat.CommentCount,
				FavoriteCount: stat.FavoriteCount,
				PublishAt:     video.CreatedAt,
			})
		}
	}

	return rows, nil
}

// ListCandidatesByIDs 按视频 ID 批量取候选快照（两段 IN 查询，不联表）。
// 已软删或缺少统计行的视频不会返回，返回顺序与入参顺序无关。
func ListCandidatesByIDs(ctx context.Context, db *gorm.DB, videoIDs []int64) ([]CandidateRow, error) {
	ids := uniquePositiveIDs(videoIDs)

	if len(ids) == 0 {
		return nil, nil
	}

	videos := make([]candidateVideo, 0, len(ids))
	for start := 0; start < len(ids); start += statQueryChunk {
		end := start + statQueryChunk
		if end > len(ids) {
			end = len(ids)
		}

		var part []candidateVideo
		if err := db.WithContext(ctx).
			Table("video_baseinfo AS v").
			Select("v.video_id, v.created_at").
			Where("v.deleted_at IS NULL").
			Where("v.video_id IN ?", ids[start:end]).
			Scan(&part).Error; err != nil {
			return nil, xerr.Wrap(err, "list hot ranking candidate videos by ids failed")
		}
		videos = append(videos, part...)
	}

	stats, err := loadStatRows(ctx, db, ids)
	if err != nil {
		return nil, err
	}

	rows := make([]CandidateRow, 0, len(videos))
	for _, video := range videos {
		stat, ok := stats[video.VideoID]
		if !ok {
			continue
		}
		rows = append(rows, CandidateRow{
			VideoID:       video.VideoID,
			VisitCount:    stat.VisitCount,
			LikeCount:     stat.LikeCount,
			CommentCount:  stat.CommentCount,
			FavoriteCount: stat.FavoriteCount,
			PublishAt:     video.CreatedAt,
		})
	}
	return rows, nil
}

// uniquePositiveIDs 去重并过滤非法 ID，减少 IN 查询的参数规模。
func uniquePositiveIDs(videoIDs []int64) []int64 {
	seen := make(map[int64]struct{}, len(videoIDs))
	ids := make([]int64, 0, len(videoIDs))
	for _, id := range videoIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

// listCandidateVideos 只查驱动表，按游标取一批视频 ID 与发布时间。
func listCandidateVideos(ctx context.Context, db *gorm.DB, since, cursorAt time.Time, cursorID int64, limit int32) ([]candidateVideo, error) {
	dbQuery := db.WithContext(ctx).
		Table("video_baseinfo AS v").
		Select("v.video_id, v.created_at").
		Where("v.deleted_at IS NULL")

	if since.IsZero() {
		// 全量模式：主键游标，避免 OFFSET 深分页。
		dbQuery = dbQuery.
			Where("v.video_id > ?", cursorID).
			Order("v.video_id ASC")
	} else {
		// 窗口模式：created_at 区间走复合索引，同毫秒用 video_id 精断。
		dbQuery = dbQuery.
			Where("v.created_at >= ?", since).
			Where("(v.created_at > ? OR (v.created_at = ? AND v.video_id > ?))", cursorAt, cursorAt, cursorID).
			Order("v.created_at ASC, v.video_id ASC")
	}

	videos := make([]candidateVideo, 0, limit)
	if err := dbQuery.Limit(int(limit)).Scan(&videos).Error; err != nil {
		return nil, xerr.Wrap(err, "list hot ranking candidate videos failed")
	}
	return videos, nil
}

// loadStatRows 按主键批量取计数快照，分块控制 IN 参数数量。
func loadStatRows(ctx context.Context, db *gorm.DB, videoIDs []int64) (map[int64]statRow, error) {
	result := make(map[int64]statRow, len(videoIDs))
	for start := 0; start < len(videoIDs); start += statQueryChunk {
		end := start + statQueryChunk
		if end > len(videoIDs) {
			end = len(videoIDs)
		}

		var rows []statRow
		if err := db.WithContext(ctx).
			Table("video_stat").
			Select("video_id, visit_count, like_count, comment_count, favorite_count").
			Where("video_id IN ?", videoIDs[start:end]).
			Scan(&rows).Error; err != nil {
			return nil, xerr.Wrap(err, "load video stat for hot ranking candidates failed")
		}
		for _, row := range rows {
			result[row.VideoID] = row
		}
	}
	return result, nil
}

// CountCandidates 统计候选数量，供 dry-run 预检使用。
// 只按驱动表计数：发布链路会同步创建 video_stat 行，缺行属于极少见的脏数据，
// 预热结果与实际重建条数可能存在的偏差可忽略，换来的是不做联表/C 表扫描。
func CountCandidates(ctx context.Context, db *gorm.DB, since time.Time) (int64, error) {
	dbQuery := db.WithContext(ctx).
		Table("video_baseinfo AS v").
		Where("v.deleted_at IS NULL")

	if !since.IsZero() {
		dbQuery = dbQuery.Where("v.created_at >= ?", since)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return 0, xerr.Wrap(err, "count hot ranking candidates failed")
	}
	return total, nil
}

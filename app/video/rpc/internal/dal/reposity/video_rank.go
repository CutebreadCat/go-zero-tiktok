package reposity

import (
	"context"
	"time"

	videoranktable "go_zero-tiktok/app/video/rpc/internal/dal/tables/video_rank"
	"go_zero-tiktok/pkg/contract"

	pkgerrors "github.com/pkg/errors"
)

// ListVideoStatsForRebuild 按候选窗口分批读取统计快照。
// since 为零值表示全量模式；否则按 (created_at, video_id) 复合游标翻页。
func (r *VideoStatRepo) ListVideoStatsForRebuild(ctx context.Context, since, cursorAt time.Time, cursorID int64, limit int32) ([]types.VideoStatSnapshot, error) {
	rows, err := videoranktable.ListCandidates(ctx, r.db, since, cursorAt, cursorID, limit)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "VideoStatRepo.ListVideoStatsForRebuild")
	}

	result := make([]types.VideoStatSnapshot, 0, len(rows))
	for _, row := range rows {
		result = append(result, types.VideoStatSnapshot{
			VideoID:       row.VideoID,
			VisitCount:    row.VisitCount,
			LikeCount:     row.LikeCount,
			CommentCount:  row.CommentCount,
			FavoriteCount: row.FavoriteCount,
			PublishAt:     row.PublishAt,
		})
	}
	return result, nil
}

// CountVideoStatsForRebuild 统计候选窗口内的视频数，供 dry-run 使用。
func (r *VideoStatRepo) CountVideoStatsForRebuild(ctx context.Context, since time.Time) (int64, error) {
	total, err := videoranktable.CountCandidates(ctx, r.db, since)
	if err != nil {
		return 0, pkgerrors.WithMessage(err, "VideoStatRepo.CountVideoStatsForRebuild")
	}
	return total, nil
}

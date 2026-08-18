package feed

import (
	"context"
	"strings"

	"go_zero-tiktok/pkg/contract"
	myutils "go_zero-tiktok/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// batchGetPopulars 根据视频列表批量查询 visit_count 等热度统计。
// video 表与 video_stat 表按 video_id 一一对应，缺失统计视为 0。
func batchGetPopulars(ctx context.Context, videos []types.VideoBaseinfo, popularRepo domainPopularRepo) []types.VideoPopular {
	if len(videos) == 0 {
		return nil
	}

	videoIDs := make([]int64, 0, len(videos))
	for _, v := range videos {
		videoIDs = append(videoIDs, v.VideoID)
	}

	populars, err := popularRepo.GetPopularVideosByIDs(ctx, videoIDs)
	if err != nil {
		logx.Errorf("batchGetPopulars failed: %v", err)
		populars = map[int64]types.VideoPopular{}
	}

	result := make([]types.VideoPopular, 0, len(videos))
	for _, v := range videos {
		result = append(result, populars[v.VideoID])
	}
	return result
}

// lastPublishedAtMs 解析视频创建时间为毫秒时间戳。
func lastPublishedAtMs(createdAt string) (int64, error) {
	createdAt = strings.TrimSpace(createdAt)
	if createdAt == "" {
		return 0, nil
	}
	t, err := myutils.StrToTime(createdAt, "")
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

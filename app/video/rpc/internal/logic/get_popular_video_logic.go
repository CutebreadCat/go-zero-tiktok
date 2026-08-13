package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetPopularVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetPopularVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPopularVideoLogic {
	return &GetPopularVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetPopularVideoLogic) GetPopularVideo(in *video_pb.GetPopularVideoRequest) (*video_pb.GetPopularVideoResponse, error) {
	if in.PageSize <= 0 || in.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	videos, videoPopulars, err := l.svcCtx.VideoService.GetPopularVideos(l.ctx, in.PageNum, in.PageSize)
	if err != nil {
		return nil, xerr.Wrap(err, "GetPopularVideo")
	}

	// 使用 Redis 缓存中的 like_count 覆盖 MySQL 中的值，保证用户看到实时计数。
	videoIDs := make([]int64, 0, len(videoPopulars))
	for _, p := range videoPopulars {
		videoIDs = append(videoIDs, p.VideoID)
	}
	likeCounts, err := l.svcCtx.InteractionService.GetLikeCounts(l.ctx, videoIDs)
	if err != nil {
		// 缓存读取失败降级为使用 MySQL 值，不影响主链路。
		logger.Warnf("GetPopularVideo get like counts from cache failed: %v", err)
		likeCounts = make(map[int64]int64, len(videoPopulars))
	}

	videoInfos := make([]*video_pb.VideoInfo, 0, len(videos))
	for _, v := range videos {
		videoInfos = append(videoInfos, &video_pb.VideoInfo{
			VideoId:     v.VideoID,
			AuthorId:    v.AuthorID,
			VideoUrl:    v.VideoURL,
			CoverUrl:    v.CoverURL,
			Title:       v.Title,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
		})
	}

	popularInfos := make([]*video_pb.VideoPopularInfo, 0, len(videoPopulars))
	for _, p := range videoPopulars {
		likeCount := p.LikeCount
		if cached, ok := likeCounts[p.VideoID]; ok {
			likeCount = cached
		}
		popularInfos = append(popularInfos, &video_pb.VideoPopularInfo{
			VideoId:       p.VideoID,
			VisitCount:    p.VisitCount,
			LikeCount:     likeCount,
			CommentCount:  p.CommentCount,
			FavoriteCount: p.FavoriteCount,
		})
	}

	return &video_pb.GetPopularVideoResponse{
		Videos:   videoInfos,
		Populars: popularInfos,
	}, nil
}

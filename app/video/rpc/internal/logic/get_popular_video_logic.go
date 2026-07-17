package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb/video_pb"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPopularVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPopularVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPopularVideoLogic {
	return &GetPopularVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
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
		popularInfos = append(popularInfos, &video_pb.VideoPopularInfo{
			VideoId:    p.VideoID,
			VisitCount: p.VisitCount,
			LikeCount:  p.LikeCount,
		})
	}

	return &video_pb.GetPopularVideoResponse{
		Videos:   videoInfos,
		Populars: popularInfos,
	}, nil
}

package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetFeedVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFeedVideoLogic {
	return &GetFeedVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetFeedVideoLogic) GetFeedVideo(in *video_pb.GetFeedVideoRequest) (*video_pb.GetFeedVideoResponse, error) {
	if in.PageSize <= 0 || in.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	videos, total, err := l.svcCtx.VideoService.GetFeedVideos(l.ctx, in.LastTime, in.PageNum, in.PageSize)
	if err != nil {
		return nil, xerr.Wrap(err, "GetFeedVideo")
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

	return &video_pb.GetFeedVideoResponse{
		Videos: videoInfos,
		Total:  total,
	}, nil
}

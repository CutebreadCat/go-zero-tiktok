package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetVideosByIDsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetVideosByIDsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVideosByIDsLogic {
	return &GetVideosByIDsLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetVideosByIDsLogic) GetVideosByIDs(in *video_pb.GetVideosByIDsRequest) (*video_pb.GetVideosByIDsResponse, error) {
	if len(in.VideoIds) == 0 {
		return &video_pb.GetVideosByIDsResponse{Videos: []*video_pb.VideoInfo{}}, nil
	}

	videos, err := l.svcCtx.VideoService.GetVideosByIDs(l.ctx, in.VideoIds)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetVideosByIDs")
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

	return &video_pb.GetVideosByIDsResponse{Videos: videoInfos}, nil
}

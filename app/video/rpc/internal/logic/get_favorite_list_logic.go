package logic

import (
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetFavoriteListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetFavoriteListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFavoriteListLogic {
	return &GetFavoriteListLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetFavoriteListLogic) GetFavoriteList(in *video_pb.GetFavoriteListRequest) (*video_pb.GetFavoriteListResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.PageSize <= 0 || in.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	videos, total, err := l.svcCtx.VideoService.GetFavoritedVideos(l.ctx, in.UserId, in.PageNum, in.PageSize)
	if err != nil {
		return nil, xerr.Wrap(err, "GetFavoriteList")
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

	return &video_pb.GetFavoriteListResponse{
		Videos: videoInfos,
		Total:  total,
	}, nil
}

package video

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type GetVideoListLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVideoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVideoListLogic {
	return &GetVideoListLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetVideoListLogic) GetVideoList(req *types.GetVideoListRequest) (resp *types.GetVideoListResponse, err error) {
	if req.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量不能超过100")
	}

	rpcResp, err := l.svcCtx.VideoRpc.GetVideoList(l.ctx, &videopb.GetVideoListRequest{
		AuthorId: req.UserId,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetVideoList.GetVideoList")
	}

	videos := make([]types.VideoBaseinfo, 0, len(rpcResp.Videos))
	for _, v := range rpcResp.Videos {
		videos = append(videos, types.VideoBaseinfo{
			VideoID:     v.VideoId,
			AuthorID:    v.AuthorId,
			VideoURL:    v.VideoUrl,
			CoverURL:    v.CoverUrl,
			Title:       v.Title,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
		})
	}

	return &types.GetVideoListResponse{
		Base:   types.BaseResponse{StatusCode: 0, StatusMsg: "查询成功"},
		Videos: videos,
		Total:  rpcResp.Total,
	}, nil
}
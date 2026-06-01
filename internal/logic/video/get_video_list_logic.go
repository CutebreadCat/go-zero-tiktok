package video

import (
	"context"

	videopb "go_zero-tiktok/app/video/rpc/video_pb/video_pb"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVideoListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVideoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVideoListLogic {
	return &GetVideoListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVideoListLogic) GetVideoList(req *types.GetVideoListRequest) (resp *types.GetVideoListResponse, err error) {
	if req.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量不能超过100")
	}

	rpcResp, err := l.svcCtx.VideoRpc.GetVideoList(l.ctx, &videopb.GetVideoListRequest{
		AuthorId: req.UserID,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
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

	resp = &types.GetVideoListResponse{
		Base:   types.BaseResponse{StatusCode: 0, StatusMsg: "查询成功"},
		Videos: videos,
	}
	if resp.Videos == nil {
		resp.Videos = []types.VideoBaseinfo{}
	}

	return resp, nil
}

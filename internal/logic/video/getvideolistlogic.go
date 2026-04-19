// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"go_zero-tiktok/internal/dal"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
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
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		return nil, xerr.New(400, "每页数量不能超过100")
	}

	videos, _, err := dal.SearchVideosByKeyword(l.ctx, "", req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
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

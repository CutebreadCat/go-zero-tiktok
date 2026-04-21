// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

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

	videos, _, err := l.svcCtx.Dal.Video.GetVideosByAuthorID(l.ctx, req.UserID, req.PageNum, req.PageSize)
	if err != nil {
		return nil, xerr.New(1002, "获取用户发布视频列表失败，请稍后重试")
	}

	resp = &types.GetVideoListResponse{
		Base:   types.BaseResponse{StatusCode: 0, StatusMsg: "查询成功"},
		Videos: videos,
	}
	if resp.Videos == nil {
		resp.Videos = []types.VideoBaseinfo{}
	}
	go func(ctx context.Context) {
		for _, video := range videos {
			if err := l.svcCtx.Dal.Popular.IncreaseVideoVisitCount(ctx, video.VideoID, 1); err != nil {
				logx.Errorf("increment visit count failed for video %s: %v", video.VideoID, err)
			}
		}

	}(l.ctx)

	return resp, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VideoPopularLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoPopularLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoPopularLogic {
	return &VideoPopularLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VideoPopularLogic) VideoPopular(req *types.VideoPopularRequest) (resp *types.VideoPopularResponse, err error) {
	videoIDs, _, err := l.svcCtx.Dal.Popular.GetPopularVideoIDsByVisitCount(l.ctx, req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}

	videos, err := l.svcCtx.Dal.Video.GetVideosByIDs(l.ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	resp = &types.VideoPopularResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		Videos: videos,
	}

	return
}

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
	videoPopulars, _, err := l.svcCtx.Dal.Popular.GetPopularVideoIDsByVisitCount(l.ctx, req.PageNum, req.PageSize)
	if err != nil {
		return nil, xerr.New(1002, "获取热门视频失败，请稍后重试")
	}
	videoIDs := make([]string, 0)
	for _, videoPopular := range videoPopulars {
		videoIDs = append(videoIDs, videoPopular.VideoID)
	}

	videos, err := l.svcCtx.Dal.Video.GetVideosByIDs(l.ctx, videoIDs)
	if err != nil {
		return nil, xerr.New(1002, "获取热门视频详情失败，请稍后重试")
	}
	Items := make([]types.Item, 0)
	for i := 0; i < len(videos); i++ {
		item := types.Item{
			Videos:        videos[i],
			VideosPopular: videoPopulars[i],
		}
		Items = append(Items, item)
	}
	resp = &types.VideoPopularResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		Videos: Items,
	}

	return
}

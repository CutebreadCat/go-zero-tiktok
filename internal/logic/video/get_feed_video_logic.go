// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFeedVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFeedVideoLogic {
	return &GetFeedVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFeedVideoLogic) GetFeedVideo(req *types.FeedVideoRequest) (resp *types.FeedVideoResponse, err error) {
	videos, total, err := l.svcCtx.VideoService.GetVideosByLastTime(l.ctx, req.LastTime, req.PageNum, req.PageSize)
	if err != nil {
		return nil, xerr.Wrap(err, "GetFeedVideo.")
	}

	items := make([]types.Item, 0, len(videos))
	for _, video := range videos {
		items = append(items, types.Item{
			Videos: video,
			VideosPopular: types.VideoPopular{
				VideoID: video.VideoID,
			},
		})
	}

	return &types.FeedVideoResponse{
		Base:   types.BaseResponse{StatusCode: 0},
		Videos: items,
		Total:  total,
	}, nil
}

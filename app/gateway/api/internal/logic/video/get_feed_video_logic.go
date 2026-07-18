// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	videopb "go_zero-tiktok/app/video/rpc/video_pb/video_pb"

	logger "go_zero-tiktok/Prometheus/logger"
)

type GetFeedVideoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFeedVideoLogic {
	return &GetFeedVideoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetFeedVideoLogic) GetFeedVideo(req *types.FeedVideoRequest) (resp *types.FeedVideoResponse, err error) {
	rpcResp, err := l.svcCtx.VideoRpc.GetFeedVideo(l.ctx, &videopb.GetFeedVideoRequest{
		LastTime: req.LastTime,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.Item, 0, len(rpcResp.Videos))
	for _, v := range rpcResp.Videos {
		items = append(items, types.Item{
			Videos: types.VideoBaseinfo{
				VideoID:     v.VideoId,
				AuthorID:    v.AuthorId,
				VideoURL:    v.VideoUrl,
				CoverURL:    v.CoverUrl,
				Title:       v.Title,
				Description: v.Description,
				CreatedAt:   v.CreatedAt,
			},
			VideosPopular: types.VideoPopular{
				VideoID: v.VideoId,
			},
		})
	}

	return &types.FeedVideoResponse{
		Base:   types.BaseResponse{StatusCode: 0},
		Videos: items,
		Total:  rpcResp.Total,
	}, nil
}

package video

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type VideoPopularLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoPopularLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoPopularLogic {
	return &VideoPopularLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *VideoPopularLogic) VideoPopular(req *types.VideoPopularRequest) (resp *types.VideoPopularResponse, err error) {
	rpcResp, err := l.svcCtx.VideoRpc.GetPopularVideo(l.ctx, &videopb.GetPopularVideoRequest{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "VideoPopular.GetPopularVideo")
	}

	items := make([]types.Item, 0, len(rpcResp.Videos))
	for i, v := range rpcResp.Videos {
		item := types.Item{
			Videos: types.VideoBaseinfo{
				VideoID:     v.VideoId,
				AuthorID:    v.AuthorId,
				VideoURL:    v.VideoUrl,
				CoverURL:    v.CoverUrl,
				Title:       v.Title,
				Description: v.Description,
				CreatedAt:   v.CreatedAt,
			},
		}
		if i < len(rpcResp.Populars) {
			p := rpcResp.Populars[i]
			item.VideosPopular = types.VideoPopular{
				VideoID:        p.VideoId,
				VisitCount:     p.VisitCount,
				LikeCount:      p.LikeCount,
				CommentCount:   p.CommentCount,
				FavoriteCount:  p.FavoriteCount,
				HotScore:       p.HotScore,
				CompletionRate: p.CompletionRate,
				StallRate:      p.StallRate,
				ErrorRate:      p.ErrorRate,
				AvgBitrateKbps: p.AvgBitrateKbps,
				AvgBufferedMs:  p.AvgBufferedMs,
				AvgStallCount:  p.AvgStallCount,
				ReportCount:    p.ReportCount,
			}
		}
		items = append(items, item)
	}

	return &types.VideoPopularResponse{
		Base:  types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		Items: items,
	}, nil
}

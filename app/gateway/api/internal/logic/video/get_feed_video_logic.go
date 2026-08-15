package video

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
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
	// 从登录态取 viewer_id 透传给 video.rpc，合并 feed:global + feed:inbox:{uid}
	// 拿不到身份（未登录）时传 0，仅返回全站候选池
	viewerID, _ := myutils.GetUserIDFromContext(l.ctx)

	rpcResp, err := l.svcCtx.VideoRpc.GetFeedVideo(l.ctx, &videopb.GetFeedVideoRequest{
		LastTime: req.LastTime,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		ViewerId: viewerID,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetFeedVideo.GetFeedVideo")
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
		Base:  types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		Total: rpcResp.Total,
		Items: items,
	}, nil
}

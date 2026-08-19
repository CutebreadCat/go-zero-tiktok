package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetLikeListLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLikeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikeListLogic {
	return &GetLikeListLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetLikeListLogic) GetLikeList(req *types.GetLikeListRequest) (resp *types.GetLikeListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	rpcResp, err := l.svcCtx.InteractionRpc.GetLikeList(l.ctx, &interactionpb.GetLikeListRequest{
		UserId:   userID,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetLikeList.GetLikeList")
	}

	videos := make([]types.VideoBaseinfo, 0, len(rpcResp.VideoIds))
	if len(rpcResp.VideoIds) > 0 {
		videoResp, err := l.svcCtx.VideoRpc.GetVideosByIDs(l.ctx, &videopb.GetVideosByIDsRequest{
			VideoIds: rpcResp.VideoIds,
		})
		if err != nil {
			return nil, xerr.HandleDaoError(err, "GetLikeList.GetVideosByIDs")
		}

		for _, v := range videoResp.Videos {
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
	}

	return &types.GetLikeListResponse{
		Base:      types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		VideoList: videos,
		LikeCount: rpcResp.Total,
	}, nil
}

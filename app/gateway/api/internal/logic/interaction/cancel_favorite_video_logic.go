package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type CancelFavoriteVideoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelFavoriteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelFavoriteVideoLogic {
	return &CancelFavoriteVideoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *CancelFavoriteVideoLogic) CancelFavoriteVideo(req *types.CancelFavoriteVideoRequest) (resp *types.CancelFavoriteVideoResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	_, err = l.svcCtx.VideoRpc.CancelFavoriteVideo(l.ctx, &videopb.CancelFavoriteVideoRequest{
		UserId:  userID,
		VideoId: req.VideoId,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "CancelFavoriteVideo.CancelFavoriteVideo")
	}

	return &types.CancelFavoriteVideoResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "取消收藏成功"},
	}, nil
}

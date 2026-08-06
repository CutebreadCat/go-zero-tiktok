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

type CancelLikeVideoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelLikeVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelLikeVideoLogic {
	return &CancelLikeVideoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *CancelLikeVideoLogic) CancelLikeVideo(req *types.CancelLikeVideoRequest) (resp *types.CancelLikeVideoResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	_, err = l.svcCtx.VideoRpc.LikeVideo(l.ctx, &videopb.LikeVideoRequest{
		UserId:     userID,
		VideoId:    req.VideoId,
		ActionType: 0, // 取消点赞
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "CancelLikeVideo.LikeVideo")
	}

	return &types.CancelLikeVideoResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "取消点赞成功"},
	}, nil
}
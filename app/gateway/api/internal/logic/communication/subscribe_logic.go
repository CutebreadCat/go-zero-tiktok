package communication

import (
	"context"

	communicationpb "go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type SubscribeLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeLogic {
	return &SubscribeLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *SubscribeLogic) Subscribe(req *types.SubscribeRequest) (resp *types.SubscribeResponse, err error) {
	followerID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.ToUserID == 0 {
		return nil, xerr.NewInvalidParam("被关注用户ID不能为空")
	}

	_, err = l.svcCtx.CommunicationRpc.Subscribe(l.ctx, &communicationpb.SubscribeRequest{
		FollowerId: followerID,
		UserId:     req.ToUserID,
		ActionType: 1, // 关注
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Subscribe.Subscribe")
	}

	return &types.SubscribeResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "关注成功"},
	}, nil
}
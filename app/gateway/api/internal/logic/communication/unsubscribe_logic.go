package communication

import (
	"context"

	communicationpb "go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type UnsubscribeLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnsubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnsubscribeLogic {
	return &UnsubscribeLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *UnsubscribeLogic) Unsubscribe(req *types.UnsubscribeRequest) (resp *types.UnsubscribeResponse, err error) {
	followerID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.ToUserID == 0 {
		return nil, xerr.NewInvalidParam("被取消关注用户ID不能为空")
	}

	_, err = l.svcCtx.CommunicationRpc.Subscribe(l.ctx, &communicationpb.SubscribeRequest{
		FollowerId: followerID,
		UserId:     req.ToUserID,
		ActionType: 0, // 取消关注
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Unsubscribe.Subscribe")
	}

	return &types.UnsubscribeResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "取消关注成功"},
	}, nil
}
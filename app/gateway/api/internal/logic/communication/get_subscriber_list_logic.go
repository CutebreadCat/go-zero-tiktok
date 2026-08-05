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

type GetSubscriberListLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSubscriberListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubscriberListLogic {
	return &GetSubscriberListLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetSubscriberListLogic) GetSubscriberList(req *types.GetSubscriberListRequest) (resp *types.GetSubscriberListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	rpcResp, err := l.svcCtx.CommunicationRpc.GetSubscriberList(l.ctx, &communicationpb.GetSubscriberListRequest{
		UserId:   userID,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetSubscriberList.GetSubscriberList")
	}

	subscriberList := make([]types.UserBaseinfo, 0, len(rpcResp.Users))
	for _, u := range rpcResp.Users {
		subscriberList = append(subscriberList, types.UserBaseinfo{
			UserID:   u.UserId,
			Username: u.Username,
			PhotoURL: u.PhotoUrl,
		})
	}

	return &types.GetSubscriberListResponse{
		Base:            types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		SubscriberList:  subscriberList,
		SubscriberCount: rpcResp.Total,
	}, nil
}
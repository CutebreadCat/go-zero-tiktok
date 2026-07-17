package communication

import (
	"context"

	communicationpb "go_zero-tiktok/app/communication/rpc/communication_pb/communication_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubscriberListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSubscriberListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubscriberListLogic {
	return &GetSubscriberListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubscriberListLogic) GetSubscriberList(req *types.GetSubscriberListRequest) (resp *types.GetSubscriberListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	rpcResp, err := l.svcCtx.CommunicationRpc.GetSubscriberList(l.ctx, &communicationpb.GetSubscriberListRequest{
		UserId:   userID,
		PageNum:  req.PageNumber,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	subscriberList := make([]types.UserBaseinfo, 0, len(rpcResp.Users))
	for _, u := range rpcResp.Users {
		subscriberList = append(subscriberList, types.UserBaseinfo{
			UserID:   u.UserId,
			Username: u.Username,
			PhotoURL: u.PhotoUrl,
		})
	}

	resp = &types.GetSubscriberListResponse{
		BaseResponse:    types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		SubscriberList:  subscriberList,
		SubscriberCount: int32(rpcResp.Total),
	}

	return resp, nil
}

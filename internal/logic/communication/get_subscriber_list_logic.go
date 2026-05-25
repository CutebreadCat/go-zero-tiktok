package communication

import (
	"context"

	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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

	subscriberList, total, err := l.svcCtx.UserFollowService.GetSubscriberList(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, err
	}

	resp = &types.GetSubscriberListResponse{
		BaseResponse:    types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		SubscriberList:  subscriberList,
		SubscriberCount: int32(total),
	}

	return resp, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

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
		return nil, err
	}

	relations, total, err := l.svcCtx.Dal.UserFollow.GetFollowingByFollowerID(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, err
	}

	subscriberIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		subscriberIDs = append(subscriberIDs, relation.UserID)
	}

	subscriberList, err := l.svcCtx.Dal.User.GetUsersByIDs(l.ctx, subscriberIDs)
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

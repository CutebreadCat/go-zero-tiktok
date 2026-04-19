// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"context"

	"go_zero-tiktok/internal/dal"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFansListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFansListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFansListLogic {
	return &GetFansListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFansListLogic) GetFansList(req *types.GetFansListRequest) (resp *types.GetFansListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	relations, total, err := dal.GetFansByUserID(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, err
	}

	fansIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		fansIDs = append(fansIDs, relation.FollowerID)
	}

	fansList, err := dal.GetUsersByIDs(l.ctx, fansIDs)
	if err != nil {
		return nil, err
	}

	resp = &types.GetFansListResponse{
		BaseResponse: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		FansList:     fansList,
		FansCount:    int32(total),
	}

	return resp, nil
}

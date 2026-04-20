// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
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
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}

	relations, total, err := l.svcCtx.Dal.UserFollow.GetFansByUserID(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, xerr.New(1002, "获取粉丝列表失败，请稍后重试")
	}

	fansIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		fansIDs = append(fansIDs, relation.FollowerID)
	}

	fansList, err := l.svcCtx.Dal.User.GetUsersByIDs(l.ctx, fansIDs)
	if err != nil {
		return nil, xerr.New(1002, "获取粉丝信息失败，请稍后重试")
	}

	resp = &types.GetFansListResponse{
		BaseResponse: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		FansList:     fansList,
		FansCount:    int32(total),
	}

	return resp, nil
}

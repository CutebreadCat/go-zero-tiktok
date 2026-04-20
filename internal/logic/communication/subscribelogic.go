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

type SubscribeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeLogic {
	return &SubscribeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubscribeLogic) Subscribe(req *types.SubscribeRequest) (resp *types.SubscribeResponse, err error) {
	followerID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}

	if req.ToUserID == "" {
		return nil, xerr.New(400, "被关注用户ID不能为空")
	}

	switch req.ActionType {
	case 1:
		err = l.svcCtx.Dal.UserFollow.FollowUser(l.ctx, followerID, req.ToUserID)
	case 0:
		err = l.svcCtx.Dal.UserFollow.UnfollowUser(l.ctx, followerID, req.ToUserID)
	default:
		return nil, xerr.New(400, "操作类型无效，仅支持1(关注)或0(取关)")
	}

	if err != nil {
		return nil, err
	}

	resp = &types.SubscribeResponse{
		BaseResponse: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
	}

	return
}

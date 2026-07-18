package user

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	myutils "go_zero-tiktok/pkg/utils"

	logger "go_zero-tiktok/Prometheus/logger"
)

type JwchGetUserCookieLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJwchGetUserCookieLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JwchGetUserCookieLogic {
	return &JwchGetUserCookieLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *JwchGetUserCookieLogic) JwchGetUserCookie(req *types.JwchGetUserCookieRequest) (resp *types.JwchGetUserCookieResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return &types.JwchGetUserCookieResponse{
			Base: types.BaseResponse{StatusCode: 400, StatusMsg: "获取用户 ID 失败"},
		}, nil
	}

	result, err := l.svcCtx.UserRpc.JwchGetUserCookie(l.ctx, &userservice.JwchGetUserCookieRequest{
		UserId: userID,
	})
	if err != nil {
		return &types.JwchGetUserCookieResponse{
			Base: types.BaseResponse{StatusCode: 400, StatusMsg: err.Error()},
		}, nil
	}
	return &types.JwchGetUserCookieResponse{
		Base:   types.BaseResponse{StatusCode: 200, StatusMsg: "获取成功"},
		JwchID: result.JwchId,
		Cookie: result.Cookie,
	}, nil
}

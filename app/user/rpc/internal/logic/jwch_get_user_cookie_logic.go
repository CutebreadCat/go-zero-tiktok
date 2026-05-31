package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type JwchGetUserCookieLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJwchGetUserCookieLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JwchGetUserCookieLogic {
	return &JwchGetUserCookieLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *JwchGetUserCookieLogic) JwchGetUserCookie(in *user_pb.JwchGetUserCookieRequest) (*user_pb.JwchGetUserCookieResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户 ID 不能为空")
	}

	identifier, cookie, err := l.svcCtx.UserJwchService.GetCookie(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &user_pb.JwchGetUserCookieResponse{
		JwchId: identifier,
		Cookie: cookie,
	}, nil
}

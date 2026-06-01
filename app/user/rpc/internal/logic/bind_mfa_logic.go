package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindMfaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindMfaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMfaLogic {
	return &BindMfaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BindMfaLogic) BindMfa(in *user_pb.BindMfaRequest) (*user_pb.BindMfaResponse, error) {
	if in.UserId == "" || in.MfaSecret == "" || in.MfaCode == "" {
		return nil, xerr.NewInvalidParam("MFA 参数不能为空")
	}

	if err := l.svcCtx.UserMfaService.BindMFA(l.ctx, in.UserId, in.MfaSecret, in.MfaCode); err != nil {
		return nil, err
	}

	return &user_pb.BindMfaResponse{}, nil
}

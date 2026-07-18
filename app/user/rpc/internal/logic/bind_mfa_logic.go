package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type BindMfaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewBindMfaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMfaLogic {
	return &BindMfaLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
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

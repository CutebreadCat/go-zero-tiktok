package logic

import (
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMfaQRCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMfaQRCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMfaQRCodeLogic {
	return &GetMfaQRCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMfaQRCodeLogic) GetMfaQRCode(in *user_pb.GetMfaQRCodeRequest) (*user_pb.GetMfaQRCodeResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户 ID 不能为空")
	}

	secret, url, err := l.svcCtx.UserMfaService.GenerateQRCode(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &user_pb.GetMfaQRCodeResponse{
		MfaSecret: secret,
		QrCodeUrl: url,
	}, nil
}

package user

import (
	"context"
	"net/http"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type GetMfaqrcodeLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMfaqrcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMfaqrcodeLogic {
	return &GetMfaqrcodeLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetMfaqrcodeLogic) GetMfaqrcode(req *types.MfaqrcodeRequest) (resp *types.MfaqrcodeResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("获取用户 ID 失败")
	}

	result, err := l.svcCtx.UserRpc.GetMfaQRCode(l.ctx, &userservice.GetMfaQRCodeRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return &types.MfaqrcodeResponse{
		Mfa_secret: result.MfaSecret,
		QRCodeURL:  result.QrCodeUrl,
		Base:       types.BaseResponse{StatusCode: http.StatusOK, StatusMsg: "获取 secret 成功"},
	}, nil
}

package user

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type GetMfaQRCodeLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMfaQRCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMfaQRCodeLogic {
	return &GetMfaQRCodeLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetMfaQRCodeLogic) GetMfaQRCode(req *types.GetMfaQRCodeRequest) (resp *types.GetMfaQRCodeResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("获取用户 ID 失败")
	}

	result, err := l.svcCtx.UserRpc.GetMfaQRCode(l.ctx, &userservice.GetMfaQRCodeRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetMfaQRCode.GetMfaQRCode")
	}

	return &types.GetMfaQRCodeResponse{
		Base:      types.BaseResponse{StatusCode: 0, StatusMsg: "获取 secret 成功"},
		QRCodeURL: result.QrCodeUrl,
		MfaSecret: result.MfaSecret,
	}, nil
}
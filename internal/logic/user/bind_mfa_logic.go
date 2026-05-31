package user

import (
	"context"

	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindMfaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindMfaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMfaLogic {
	return &BindMfaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindMfaLogic) BindMfa(req *types.BindMfaqrcodeRequest) (resp *types.BindMfaqrcodeResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("获取用户 ID 失败")
	}

	if l.svcCtx.UserRpc != nil {
		if _, err := l.svcCtx.UserRpc.BindMfa(l.ctx, &userservice.BindMfaRequest{
			UserId:    userID,
			MfaSecret: req.Mfa_secret,
			MfaCode:   req.Mfa_code,
		}); err != nil {
			return nil, err
		}
		return &types.BindMfaqrcodeResponse{
			Base: types.BaseResponse{StatusCode: 0, StatusMsg: "绑定 MFA 成功"},
		}, nil
	}

	if err := l.svcCtx.UserMfaService.BindMFA(l.ctx, userID, req.Mfa_secret, req.Mfa_code); err != nil {
		return nil, err
	}

	return &types.BindMfaqrcodeResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "绑定 MFA 成功"},
	}, nil
}

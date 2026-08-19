package user

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type BindMfaLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindMfaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMfaLogic {
	return &BindMfaLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *BindMfaLogic) BindMfa(req *types.BindMfaRequest) (resp *types.BindMfaResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("获取用户 ID 失败")
	}

	if _, err := l.svcCtx.UserRpc.BindMfa(l.ctx, &userservice.BindMfaRequest{
		UserId:    userID,
		MfaSecret: req.MfaSecret,
		MfaCode:   req.MfaCode,
	}); err != nil {
		return nil, xerr.HandleDaoError(err, "BindMfa.BindMfa")
	}

	return &types.BindMfaResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "绑定 MFA 成功"},
	}, nil
}

package user

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type RegisterLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, xerr.NewInvalidParam("用户名或密码不能为空")
	}

	result, err := l.svcCtx.UserRpc.Register(l.ctx, &userservice.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Register.Register")
	}

	return &types.RegisterResponse{
		Base:   types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		UserID: result.UserId,
	}, nil
}
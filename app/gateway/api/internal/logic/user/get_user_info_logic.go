package user

import (
	"context"

	token "go_zero-tiktok/app/gateway/api/internal/middleware/token"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetUserInfoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoRequest) (resp *types.GetUserInfoResponse, err error) {
	userID := token.UserIDFromContext(l.ctx)
	if userID == 0 {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	result, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userservice.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetUserInfo.GetUserInfo")
	}

	return &types.GetUserInfoResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		User: types.UserBaseinfo{
			UserID:    result.User.UserId,
			Username:  result.User.Username,
			PhotoURL:  result.User.PhotoUrl,
			CreatedAt: result.User.CreatedAt,
			UpdatedAt: result.User.UpdatedAt,
		},
	}, nil
}

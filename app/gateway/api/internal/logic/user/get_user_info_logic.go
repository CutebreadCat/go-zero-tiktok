package user

import (
	"context"

	token "go_zero-tiktok/app/gateway/api/internal/middleware/token"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
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

func (l *GetUserInfoLogic) GetUserInfo(req *types.UserInfoRequest) (*types.UserInfoResponse, error) {
	userID := token.UserIDFromContext(l.ctx)
	if userID == "" {
		userID = req.UserID
	}
	if userID == "" {
		return nil, xerr.NewInvalidParam("用户 ID 不能为空")
	}

	result, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userservice.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	return &types.UserInfoResponse{
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

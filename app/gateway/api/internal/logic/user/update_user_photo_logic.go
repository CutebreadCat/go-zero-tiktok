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

type UpdateUserPhotoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserPhotoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserPhotoLogic {
	return &UpdateUserPhotoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *UpdateUserPhotoLogic) UpdateUserPhoto(photo []byte) (resp *types.UpdateUserPhotoResponse, err error) {
	userID := token.UserIDFromContext(l.ctx)
	if userID == 0 {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	result, err := l.svcCtx.UserRpc.UpdateUserPhoto(l.ctx, &userservice.UpdateUserPhotoRequest{
		UserId: userID,
		Photo:  photo,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "UpdateUserPhoto.UpdateUserPhoto")
	}

	return &types.UpdateUserPhotoResponse{
		Base:     types.BaseResponse{StatusCode: 0, StatusMsg: "照片上传成功"},
		PhotoURL: result.PhotoUrl,
	}, nil
}

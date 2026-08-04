package user

import (
	"context"
	"io"
	"mime/multipart"

	token "go_zero-tiktok/app/gateway/api/internal/middleware/token"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type PostUserPhotoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPostUserPhotoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostUserPhotoLogic {
	return &PostUserPhotoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *PostUserPhotoLogic) PostUserPhoto(req *types.UserphotoRequest, file multipart.File) (resp *types.UserphotoResponse, err error) {
	userID := token.UserIDFromContext(l.ctx)
	if userID == 0 {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	photo, err := io.ReadAll(file)
	if err != nil {
		return nil, xerr.Wrap(err, "PostUserPhoto.ReadAll")
	}
	if _, err := l.svcCtx.UserRpc.UpdateUserPhoto(l.ctx, &userservice.UpdateUserPhotoRequest{
		UserId: userID,
		Photo:  photo,
	}); err != nil {
		return nil, err
	}
	return &types.UserphotoResponse{
		StatusCode: 200,
		StatusMsg:  "照片上传成功",
	}, nil
}

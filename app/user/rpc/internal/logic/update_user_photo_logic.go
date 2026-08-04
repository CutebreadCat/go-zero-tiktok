package logic

import (
	"bytes"
	"context"

	"go_zero-tiktok/app/user/rpc/internal/svc"
	"go_zero-tiktok/app/user/rpc/user_pb"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type UpdateUserPhotoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewUpdateUserPhotoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserPhotoLogic {
	return &UpdateUserPhotoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *UpdateUserPhotoLogic) UpdateUserPhoto(in *user_pb.UpdateUserPhotoRequest) (*user_pb.UpdateUserPhotoResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户 ID 不能为空")
	}
	if len(in.Photo) == 0 {
		return nil, xerr.NewInvalidParam("头像文件不能为空")
	}

	photoURL, err := l.svcCtx.UserProfileService.UpdatePhoto(l.ctx, in.UserId, bytes.NewReader(in.Photo))
	if err != nil {
		return nil, err
	}

	return &user_pb.UpdateUserPhotoResponse{
		PhotoUrl: photoURL,
	}, nil
}

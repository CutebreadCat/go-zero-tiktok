package logic

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/communication/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type CountUnreadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewCountUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CountUnreadLogic {
	return &CountUnreadLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *CountUnreadLogic) CountUnread(in *communication_pb.CountUnreadRequest) (*communication_pb.CountUnreadResponse, error) {
	count, err := l.svcCtx.MessageService.CountUnread(l.ctx, in.UserId)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "CountUnread")
	}
	return &communication_pb.CountUnreadResponse{UnreadCount: count}, nil
}

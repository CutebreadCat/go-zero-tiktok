package logic

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/communication/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type MarkReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *MarkReadLogic) MarkRead(in *communication_pb.MarkReadRequest) (*communication_pb.MarkReadResponse, error) {
	count, err := l.svcCtx.MessageService.MarkRead(l.ctx, in.UserId, in.MessageIds)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "MarkRead")
	}
	return &communication_pb.MarkReadResponse{UpdatedCount: count}, nil
}

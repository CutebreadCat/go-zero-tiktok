package logic

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/communication/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type CreateMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewCreateMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMessageLogic {
	return &CreateMessageLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *CreateMessageLogic) CreateMessage(in *communication_pb.CreateMessageRequest) (*communication_pb.CreateMessageResponse, error) {
	result, err := l.svcCtx.MessageService.Create(
		l.ctx,
		in.ReceiverId,
		in.Type,
		in.Title,
		in.Content,
		in.EventId,
		"", // idempotencyKey 由调用方通过 metadata 传入，logic 层暂不处理
		in.SenderId,
		in.SenderNickname,
		in.SenderAvatarUrl,
		in.TargetId,
		in.TargetType,
	)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "CreateMessage")
	}

	resp := &communication_pb.CreateMessageResponse{Created: result.Created}
	if result.Message != nil {
		resp.MessageId = result.Message.ID
	}
	return resp, nil
}

package logic

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb"
	domainmessage "go_zero-tiktok/app/communication/rpc/internal/domain/message"
	"go_zero-tiktok/app/communication/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessagesLogic {
	return &GetMessagesLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetMessagesLogic) GetMessages(in *communication_pb.GetMessagesRequest) (*communication_pb.GetMessagesResponse, error) {
	result, err := l.svcCtx.MessageService.List(l.ctx, in.UserId, in.Type, in.Cursor, int(in.Limit))
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetMessages")
	}

	items := make([]*communication_pb.MessageInfo, 0, len(result.Items))
	for _, msg := range result.Items {
		items = append(items, messageToPb(msg))
	}

	return &communication_pb.GetMessagesResponse{
		Items:      items,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}, nil
}

func messageToPb(msg *domainmessage.Message) *communication_pb.MessageInfo {
	if msg == nil {
		return nil
	}
	info := &communication_pb.MessageInfo{
		MessageId:        msg.ID,
		ReceiverId:       msg.ReceiverID,
		Type:             msg.Type,
		Title:            msg.Title,
		Content:          msg.Content,
		EventId:          msg.EventID,
		SenderId:         msg.SenderID,
		SenderNickname:   msg.SenderNickname,
		SenderAvatarUrl:  msg.SenderAvatarURL,
		TargetId:         msg.TargetID,
		TargetType:       msg.TargetType,
		IsRead:           msg.IsRead,
		CreatedAt:        msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if msg.ReadAt != nil {
		info.ReadAt = msg.ReadAt.Format("2006-01-02T15:04:05Z")
	}
	return info
}

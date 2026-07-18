package logic

import (
	"context"

	"go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/chat/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
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

func (l *GetMessagesLogic) GetMessages(in *chat_pb.GetMessagesRequest) (*chat_pb.GetMessagesResponse, error) {
	if in.RoomId == "" {
		return nil, xerr.NewInvalidParam("房间ID不能为空")
	}
	if in.PageSize <= 0 || in.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	messages, err := l.svcCtx.ChatService.GetMessages(l.ctx, in.RoomId, int(in.PageSize), int(in.PageNum))
	if err != nil {
		return nil, xerr.Wrap(err, "GetMessages")
	}

	msgInfos := make([]*chat_pb.MessageInfo, 0, len(messages))
	for _, msg := range messages {
		if msg != nil {
			msgInfos = append(msgInfos, &chat_pb.MessageInfo{
				MessageId: msg.ID,
				RoomId:    msg.RoomID,
				SenderId:  msg.SenderID,
				Content:   msg.Content,
				CreatedAt: msg.CreatedAt,
			})
		}
	}

	return &chat_pb.GetMessagesResponse{
		Messages: msgInfos,
	}, nil
}

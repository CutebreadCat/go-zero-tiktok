package logic

import (
	"context"

	"go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/chat/rpc/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StoreChatMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStoreChatMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StoreChatMessageLogic {
	return &StoreChatMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StoreChatMessageLogic) StoreChatMessage(in *chat_pb.StoreChatMessageRequest) (*chat_pb.StoreChatMessageResponse, error) {
	if in.RoomId == "" {
		return nil, xerr.NewInvalidParam("房间ID不能为空")
	}
	if in.SenderId == "" {
		return nil, xerr.NewInvalidParam("发送者ID不能为空")
	}

	message := &types.MessageChat{
		ID:        in.MessageId,
		RoomID:    in.RoomId,
		SenderID:  in.SenderId,
		Content:   in.Content,
		CreatedAt: in.CreatedAt,
	}

	if err := l.svcCtx.ChatService.StoreChatMessage(l.ctx, message); err != nil {
		return nil, xerr.Wrap(err, "StoreChatMessage")
	}

	return &chat_pb.StoreChatMessageResponse{}, nil
}

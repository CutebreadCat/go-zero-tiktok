package logic

import (
	"context"

	"go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/chat/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type GetChatRoomsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetChatRoomsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatRoomsLogic {
	return &GetChatRoomsLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetChatRoomsLogic) GetChatRooms(in *chat_pb.GetChatRoomsRequest) (*chat_pb.GetChatRoomsResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}

	roomIDs, err := l.svcCtx.ChatService.GetJoinRooms(l.ctx, in.UserId)
	if err != nil {
		return nil, xerr.Wrap(err, "GetChatRooms")
	}

	return &chat_pb.GetChatRoomsResponse{
		RoomIds: roomIDs,
	}, nil
}

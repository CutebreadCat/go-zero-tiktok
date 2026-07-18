package logic

import (
	"context"

	"go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/chat/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type JoinChatRoomLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewJoinChatRoomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinChatRoomLogic {
	return &JoinChatRoomLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *JoinChatRoomLogic) JoinChatRoom(in *chat_pb.JoinChatRoomRequest) (*chat_pb.JoinChatRoomResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.RoomId == "" {
		return nil, xerr.NewInvalidParam("房间ID不能为空")
	}

	if err := l.svcCtx.ChatService.JoinChatRoom(l.ctx, in.UserId, in.RoomId); err != nil {
		return nil, xerr.HandleDaoError(err, "JoinChatRoom")
	}

	return &chat_pb.JoinChatRoomResponse{}, nil
}

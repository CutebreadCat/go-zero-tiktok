package logic

import (
	"context"

	"go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/chat/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type JoinChatRoomLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJoinChatRoomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinChatRoomLogic {
	return &JoinChatRoomLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
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

package logic

import (
	"context"

	"go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/chat/rpc/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatRoomUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatRoomUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatRoomUsersLogic {
	return &GetChatRoomUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetChatRoomUsersLogic) GetChatRoomUsers(in *chat_pb.GetChatRoomUsersRequest) (*chat_pb.GetChatRoomUsersResponse, error) {
	if in.RoomId == "" {
		return nil, xerr.NewInvalidParam("房间ID不能为空")
	}

	userIDs, err := l.svcCtx.ChatService.GetChatRoomUsers(l.ctx, in.RoomId)
	if err != nil {
		return nil, xerr.Wrap(err, "GetChatRoomUsers")
	}

	return &chat_pb.GetChatRoomUsersResponse{
		UserIds: userIDs,
	}, nil
}

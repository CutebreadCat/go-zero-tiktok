package logic

import (
	"context"
	"strings"

	"go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/chat/rpc/internal/svc"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatRoomLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatRoomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatRoomLogic {
	return &CreateChatRoomLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateChatRoomLogic) CreateChatRoom(in *chat_pb.CreateChatRoomRequest) (*chat_pb.CreateChatRoomResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.Type != 0 && in.Type != 1 {
		return nil, xerr.NewInvalidParam("聊天室类型无效")
	}
	if in.Type == 1 && strings.TrimSpace(in.RoomName) == "" {
		return nil, xerr.NewInvalidParam("群聊名称不能为空")
	}
	if in.Type == 0 && len(in.UserIds) == 0 {
		return nil, xerr.NewInvalidParam("私聊至少需要一个对方用户")
	}

	roomID := myutils.GenerateRoomID()
	userSet := map[string]struct{}{in.UserId: {}}
	for _, uid := range in.UserIds {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		userSet[uid] = struct{}{}
	}
	userIDs := make([]string, 0, len(userSet))
	for uid := range userSet {
		userIDs = append(userIDs, uid)
	}

	if err := l.svcCtx.ChatService.CreateChatRoom(l.ctx, roomID, in.Type, in.RoomName, userIDs); err != nil {
		return nil, xerr.HandleDaoError(err, "CreateChatRoom")
	}

	return &chat_pb.CreateChatRoomResponse{
		RoomId: roomID,
	}, nil
}

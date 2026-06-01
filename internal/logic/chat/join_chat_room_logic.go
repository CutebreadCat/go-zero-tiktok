package chat

import (
	"context"

	chatpb "go_zero-tiktok/app/chat/rpc/chat_pb/chat_pb"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type JoinChatRoomLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJoinChatRoomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinChatRoomLogic {
	return &JoinChatRoomLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JoinChatRoomLogic) JoinChatRoom(req *types.JoinChatRoomRequest) (resp *types.JoinChatRoomResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.RoomID == "" {
		return nil, xerr.NewInvalidParam("聊天室ID不能为空")
	}

	_, err = l.svcCtx.ChatRpc.JoinChatRoom(l.ctx, &chatpb.JoinChatRoomRequest{
		UserId: userID,
		RoomId: req.RoomID,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.JoinChatRoomResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
	}
	return resp, nil
}

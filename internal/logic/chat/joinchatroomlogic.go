// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
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
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}
	if req.RoomID == "" {
		return nil, xerr.New(400, "聊天室ID不能为空")
	}
	if err := l.svcCtx.Dal.Chat.JoinChatRoom(l.ctx, userID, req.RoomID); err != nil {
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

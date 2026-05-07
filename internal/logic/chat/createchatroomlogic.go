// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"strings"

	chattable "go_zero-tiktok/internal/dal/tables/chat"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatRoomLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChatRoomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatRoomLogic {
	return &CreateChatRoomLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatRoomLogic) CreateChatRoom(req *types.CreateChatRoomRequest) (resp *types.CreateChatRoomResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}
	if req.Types != 0 && req.Types != 1 {
		return nil, xerr.New(400, "聊天室类型无效")
	}
	if req.Types == 1 && strings.TrimSpace(req.RoomName) == "" {
		return nil, xerr.New(400, "聊天室名称不能为空")
	}
	if req.Types == 0 && len(req.UserIDs) == 0 {
		return nil, xerr.New(400, "私聊至少需要一个对方用户")
	}

	roomID := myutils.GenerateRoomID()
	userSet := map[string]struct{}{userID: {}}
	for _, uid := range req.UserIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		userSet[uid] = struct{}{}
	}
	for uid := range userSet {
		chatRow := &types.User_chat{
			UserID:   uid,
			RoomID:   roomID,
			Leix:     req.Types,
			RoomName: req.RoomName,
		}
		if err := chattable.CreateChatRoom(l.ctx, l.svcCtx.DB, chatRow); err != nil {
			return nil, err
		}
	}

	resp = &types.CreateChatRoomResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		RoomID: roomID,
	}
	return resp, nil
}

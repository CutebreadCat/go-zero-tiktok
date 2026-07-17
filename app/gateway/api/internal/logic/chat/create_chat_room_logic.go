package chat

import (
	"context"
	"strings"

	chatpb "go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

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
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.Types != 0 && req.Types != 1 {
		return nil, xerr.NewInvalidParam("聊天室类型无效")
	}
	if req.Types == 1 && strings.TrimSpace(req.RoomName) == "" {
		return nil, xerr.NewInvalidParam("聊天室名称不能为空")
	}
	if req.Types == 0 && len(req.UserIDs) == 0 {
		return nil, xerr.NewInvalidParam("私聊至少需要一个对方用户")
	}

	userSet := map[string]struct{}{userID: {}}
	for _, uid := range req.UserIDs {
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

	rpcResp, err := l.svcCtx.ChatRpc.CreateChatRoom(l.ctx, &chatpb.CreateChatRoomRequest{
		UserId:   userID,
		Type:     req.Types,
		RoomName: req.RoomName,
		UserIds:  userIDs,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.CreateChatRoomResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		RoomID: rpcResp.RoomId,
	}
	return resp, nil
}

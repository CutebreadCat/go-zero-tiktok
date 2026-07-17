package chat

import (
	"context"

	chatpb "go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatRoomsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChatRoomsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatRoomsLogic {
	return &GetChatRoomsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatRoomsLogic) GetChatRooms(req *types.GetChatRoomsRequest) (resp *types.GetChatRoomsResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	rpcResp, err := l.svcCtx.ChatRpc.GetChatRooms(l.ctx, &chatpb.GetChatRoomsRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetChatRoomsResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		RoomsId: rpcResp.RoomIds,
	}
	return resp, nil
}

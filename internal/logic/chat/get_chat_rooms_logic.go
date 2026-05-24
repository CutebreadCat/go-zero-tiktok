package chat

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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
	rooms, err := l.svcCtx.Dal.Chat.GetJoinRooms(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	resp = &types.GetChatRoomsResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		RoomsId: rooms,
	}
	return resp, nil
}

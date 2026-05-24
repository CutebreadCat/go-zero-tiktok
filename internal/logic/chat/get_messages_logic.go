package chat

import (
	"context"

	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessagesLogic {
	return &GetMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessagesLogic) GetMessages(req *types.GetMessagesRequest) (resp *types.GetMessagesResponse, err error) {
	if req.RoomID == "" {
		return nil, xerr.NewInvalidParam("聊天室ID不能为空")
	}
	pageSize := int(req.PageSize)
	pageNumber := int(req.PageNumber)
	messages, err := l.svcCtx.Dal.Chat.GetChatRoomMessage(l.ctx, req.RoomID, pageSize, pageNumber)
	if err != nil {
		return nil, err
	}

	msgList := make([]types.MessageChat, 0, len(messages))
	for _, msg := range messages {
		if msg != nil {
			msgList = append(msgList, *msg)
		}
	}
	resp = &types.GetMessagesResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		Messages: msgList,
	}

	return resp, nil
}

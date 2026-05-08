// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
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
		return nil, xerr.New(400, "聊天室ID不能为空")
	}
	pageSize := int(req.PageSize)
	pageNumber := int(req.PageNumber)
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageNumber <= 0 {
		pageNumber = 1
	}

	messages, err := l.svcCtx.Dal.Chat.GetChatRoomMessage(l.ctx, req.RoomID, pageSize, pageNumber)
	if err != nil {
		return nil, err
	}

	resp = &types.GetMessagesResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		Messages: make([]types.MessageChat, 0, len(messages)),
	}
	for _, message := range messages {
		if message == nil {
			continue
		}
		resp.Messages = append(resp.Messages, *message)
	}

	return resp, nil
}

package chat

import (
	"context"

	chatpb "go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	"go_zero-tiktok/pkg/xerr"

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

	rpcResp, err := l.svcCtx.ChatRpc.GetMessages(l.ctx, &chatpb.GetMessagesRequest{
		RoomId:   req.RoomID,
		PageNum:  req.PageNumber,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	msgList := make([]types.MessageChat, 0, len(rpcResp.Messages))
	for _, msg := range rpcResp.Messages {
		msgList = append(msgList, types.MessageChat{
			ID:        msg.MessageId,
			RoomID:    msg.RoomId,
			SenderID:  msg.SenderId,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
		})
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

package svc

import (
	"context"

	chatpb "go_zero-tiktok/app/chat/rpc/chat_pb"
	"go_zero-tiktok/internal/types"
)

// ChatRepoAdapter WebSocket Hub 需要的 Chat 仓储适配器
type ChatRepoAdapter struct {
	chatRpc chatpb.ChatServiceClient
}

func NewChatRepoAdapter(chatRpc chatpb.ChatServiceClient) *ChatRepoAdapter {
	return &ChatRepoAdapter{chatRpc: chatRpc}
}

// GetJoinRooms 实现 websocket.RoomRepository 接口
func (a *ChatRepoAdapter) GetJoinRooms(ctx context.Context, userID string) ([]string, error) {
	resp, err := a.chatRpc.GetChatRooms(ctx, &chatpb.GetChatRoomsRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return resp.RoomIds, nil
}

// GetChatRoomUsers 实现 websocket.RoomRepository 接口
func (a *ChatRepoAdapter) GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	resp, err := a.chatRpc.GetChatRoomUsers(ctx, &chatpb.GetChatRoomUsersRequest{
		RoomId: roomID,
	})
	if err != nil {
		return nil, err
	}
	return resp.UserIds, nil
}

// StoreChatMessage 实现 websocket.MessageRepository 接口
func (a *ChatRepoAdapter) StoreChatMessage(ctx context.Context, message *types.MessageChat) error {
	_, err := a.chatRpc.StoreChatMessage(ctx, &chatpb.StoreChatMessageRequest{
		MessageId: message.ID,
		RoomId:    message.RoomID,
		SenderId:  message.SenderID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	})
	return err
}

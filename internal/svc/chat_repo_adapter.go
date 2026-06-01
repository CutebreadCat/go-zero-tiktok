package svc

import (
	"context"

	chatpb "go_zero-tiktok/app/chat/rpc/chat_pb/chat_pb"
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
	// 注意：当前 proto 中没有定义 GetChatRoomUsers 方法
	// 这里暂时返回空数组，后续可以在 proto 中添加该方法
	return []string{}, nil
}

// StoreChatMessage 实现 websocket.MessageRepository 接口
func (a *ChatRepoAdapter) StoreChatMessage(ctx context.Context, message *types.MessageChat) error {
	// 注意：当前 proto 中没有定义 StoreChatMessage 方法
	// 这里暂时返回 nil，后续可以在 proto 中添加该方法
	return nil
}

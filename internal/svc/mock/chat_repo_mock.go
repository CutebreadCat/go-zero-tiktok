package mock

import (
	"context"

	"go_zero-tiktok/internal/types"
)

type ChatRepo struct {
	CreateChatRoomFromParamsFn func(ctx context.Context, userID, roomID string, leix int32, roomName string) error
	GetChatRoomMessageFn      func(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*types.MessageChat, error)
	GetJoinRoomsFn            func(ctx context.Context, userID string) ([]string, error)
	JoinChatRoomFn            func(ctx context.Context, userID string, roomID string) error
	GetChatRoomUsersFn        func(ctx context.Context, roomID string) ([]string, error)
	StoreChatMessageFn        func(ctx context.Context, message *types.MessageChat) error
}

func (m *ChatRepo) CreateChatRoomFromParams(ctx context.Context, userID, roomID string, leix int32, roomName string) error {
	if m.CreateChatRoomFromParamsFn != nil {
		return m.CreateChatRoomFromParamsFn(ctx, userID, roomID, leix, roomName)
	}
	return nil
}

func (m *ChatRepo) GetChatRoomMessage(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*types.MessageChat, error) {
	if m.GetChatRoomMessageFn != nil {
		return m.GetChatRoomMessageFn(ctx, roomID, pageSize, pageNum)
	}
	return nil, nil
}

func (m *ChatRepo) GetJoinRooms(ctx context.Context, userID string) ([]string, error) {
	if m.GetJoinRoomsFn != nil {
		return m.GetJoinRoomsFn(ctx, userID)
	}
	return nil, nil
}

func (m *ChatRepo) JoinChatRoom(ctx context.Context, userID string, roomID string) error {
	if m.JoinChatRoomFn != nil {
		return m.JoinChatRoomFn(ctx, userID, roomID)
	}
	return nil
}

func (m *ChatRepo) GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	if m.GetChatRoomUsersFn != nil {
		return m.GetChatRoomUsersFn(ctx, roomID)
	}
	return nil, nil
}

func (m *ChatRepo) StoreChatMessage(ctx context.Context, message *types.MessageChat) error {
	if m.StoreChatMessageFn != nil {
		return m.StoreChatMessageFn(ctx, message)
	}
	return nil
}

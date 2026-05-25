package chat

import (
	"context"
	"go_zero-tiktok/internal/types"
)

type IChatRepo interface {
	CreateChatRoomFromParams(ctx context.Context, userID, roomID string, leix int32, roomName string) error
	GetChatRoomMessage(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*types.MessageChat, error)
	GetJoinRooms(ctx context.Context, userID string) ([]string, error)
	JoinChatRoom(ctx context.Context, userID string, roomID string) error
	GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error)
	StoreChatMessage(ctx context.Context, message *types.MessageChat) error
}

package websocket

import "context"

// ==================== 接口定义 ====================

// RoomRepository 房间数据仓储接口
type RoomRepository interface {
	GetJoinRooms(ctx context.Context, userID string) ([]string, error)
	GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error)
}

// RoomCache 房间缓存接口
type RoomCache interface {
	JoinRoom(ctx context.Context, roomID string, userID string) error
	LeaveRoom(ctx context.Context, roomID string, userID string) error
	RoomHeartBeat(ctx context.Context, roomID string) error
}

// RoomManager 房间管理接口
type RoomManager interface {
	LoadRooms(ctx context.Context, client *Client)
	IsMember(client *Client, roomID string) bool
	RemoveFromRooms(ctx context.Context, client *Client)
	BroadcastToRoom(roomID string, message any)
	IsOnline(userID string) bool
}

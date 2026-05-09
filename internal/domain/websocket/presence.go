package websocket

import (
	"context"
	"sync"
)

// PresenceCache 是上线状态的缓存依赖抽象
type PresenceCache interface {
	SetOnline(ctx context.Context, userID string, addr string) error
	SetOffline(ctx context.Context, userID string, addr string) error
	HeartBeat(ctx context.Context, userID string, addr string) error
}

// RoomCache 是房间缓存的依赖抽象
type RoomCache interface {
	JoinRoom(ctx context.Context, roomID string, userID string) error
	LeaveRoom(ctx context.Context, roomID string, userID string) error
	RoomHeartBeat(ctx context.Context, roomID string) error
}

// PresenceManager 管理客户端连接的上线/下线和注册
type PresenceManager interface {
	AddClient(ctx context.Context, client *Client)
	RemoveClient(ctx context.Context, client *Client)
}

// presenceManager 是 PresenceManager 的实现
type presenceManager struct {
	mu      sync.RWMutex
	clients map[*Client]bool
	cache   PresenceCache
	rooms   RoomManager
}

func (pm *presenceManager) AddClient(ctx context.Context, client *Client) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.clients[client] = true
	pm.cache.SetOnline(ctx, client.UserID, client.Conn.RemoteAddr().String())
	pm.rooms.LoadRooms(ctx, client)
}

func (pm *presenceManager) RemoveClient(ctx context.Context, client *Client) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.clients, client)
	pm.cache.SetOffline(ctx, client.UserID, client.Conn.RemoteAddr().String())
	pm.rooms.RemoveFromRooms(ctx, client)
	close(client.Send)
}

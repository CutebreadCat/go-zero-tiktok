package websocket

import (
	"context"
	"sync"
)

// ==================== PresenceManager 实现 ====================

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
	_ = pm.cache.SetOnline(ctx, client.UserID, client.Conn.RemoteAddr().String())
	pm.rooms.LoadRooms(ctx, client)
}

func (pm *presenceManager) RemoveClient(ctx context.Context, client *Client) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.clients, client)
	_ = pm.cache.SetOffline(ctx, client.UserID, client.Conn.RemoteAddr().String())
	pm.rooms.RemoveFromRooms(ctx, client)
	close(client.Send)
}

package websocket

import (
	"context"
	"log"
	"sync"
)

type roomManager struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]bool
	repo  RoomRepository
	cache RoomCache
}

func (rm *roomManager) LoadRooms(ctx context.Context, client *Client) {
	rooms, err := rm.repo.GetJoinRooms(ctx, client.UserID)
	if err != nil {
		log.Printf("加载用户 %s 的房间失败: %v", client.UserID, err)
		return
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for _, roomID := range rooms {
		client.Rooms[roomID] = true
		if rm.rooms[roomID] == nil {
			rm.rooms[roomID] = make(map[*Client]bool)
		}
		rm.rooms[roomID][client] = true
		_ = rm.cache.JoinRoom(ctx, roomID, client.UserID)
	}
}

func (rm *roomManager) IsMember(client *Client, roomID string) bool {
	return client.Rooms[roomID]
}

func (rm *roomManager) RemoveFromRooms(ctx context.Context, client *Client) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for roomID := range client.Rooms {
		if clientsInRoom, ok := rm.rooms[roomID]; ok {
			delete(clientsInRoom, client)
			if len(clientsInRoom) == 0 {
				delete(rm.rooms, roomID)
			}
		}
		_ = rm.cache.LeaveRoom(ctx, roomID, client.UserID)
	}
}

func (rm *roomManager) BroadcastToRoom(roomID string, message any) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if clientsInRoom, ok := rm.rooms[roomID]; ok {
		for client := range clientsInRoom {
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}

func (rm *roomManager) IsOnline(userID string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	for _, clientsInRoom := range rm.rooms {
		for client := range clientsInRoom {
			if client.UserID == userID {
				return true
			}
		}
	}
	return false
}

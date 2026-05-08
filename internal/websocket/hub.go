package websocket

import (
	"context"
)

func (h *Hub) AddClient(ctx context.Context, client *Client) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.Clients[client] = true
	h.Cache.SetOnline(ctx, client.UserID, client.Conn.RemoteAddr().String())
	client.LoadRooms(ctx)
}

func (h *Hub) RemoveClient(ctx context.Context, client *Client) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	delete(h.Clients, client)
	h.Cache.SetOffline(ctx, client.UserID, client.Conn.RemoteAddr().String())
	for roomID := range client.Rooms {
		if clientsInRoom, ok := h.Rooms[roomID]; ok {
			delete(clientsInRoom, client)
			if len(clientsInRoom) == 0 {
				delete(h.Rooms, roomID)
			}
		}
	}
	close(client.Send)
}

func (h *Hub) BroadcastToRoom(roomID string, message any) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()
	if clientsInRoom, ok := h.Rooms[roomID]; ok {
		for client := range clientsInRoom {
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}

package websocket

import (
	"context"
	"log"
)

func (c *Client) LoadRooms(ctx context.Context) {
	rooms, err := c.Hub.Chat.GetJoinRooms(ctx, c.UserID)
	if err != nil {
		log.Printf("Failed to load rooms for user %s: %v", c.UserID, err)
		return
	}
	for _, roomID := range rooms {
		c.Rooms[roomID] = true
		if c.Hub.Rooms[roomID] == nil {
			c.Hub.Rooms[roomID] = make(map[*Client]bool)
		}
		c.Hub.Rooms[roomID][c] = true
	}
}

func (c *Client) IsMember(roomID string) bool {
	return c.Rooms[roomID]
}

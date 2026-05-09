package websocket

import (
	"context"
	"log"
	"time"

	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	ws "github.com/gorilla/websocket"
)

type Message struct {
	message types.MessageChat
	Typek   string `json:"typek"`
}

var pongWait = 60 * time.Second

func (c *Client) ReadLoop() {
	ctx, hand := context.WithTimeout(context.Background(), time.Hour*24)
	defer func() {
		c.Hub.RemoveClient(ctx, c)
		c.Conn.Close()
		hand()
	}()
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var msg Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			break
		}
		msg.message.CreatedAt = myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05")
		switch msg.Typek {
		case "message":
			c.handleMessage(ctx, &msg.message)
		case "ping":
			c.Send <- map[string]string{"req": "pong"}
		default:
			log.Printf("Unknown message type: %s", msg.Typek)
			continue
		}
	}
}

func (c *Client) WriteLoop() {
	ticker := time.NewTicker(pongWait / 2)
	defer ticker.Stop()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(pongWait))
			c.Cmu.Lock()
			if !ok {
				log.Println("Send channel closed, exiting WriteLoop")
				c.Cmu.Unlock()
				return
			}
			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("Failed to write message to client %s: %v", c.UserID, err)
				c.Cmu.Unlock()
				return
			}
			c.Cmu.Unlock()
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(pongWait))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping to client %s: %v", c.UserID, err)
				return
			}
		}
	}
}

func (c *Client) handleMessage(ctx context.Context, msg *types.MessageChat) {
	if !c.IsMember(msg.RoomID) {
		log.Printf("User %s is not a member of room %s, ignoring message", c.UserID, msg.RoomID)
		return
	}

	go func() {
		if _, err := c.Hub.Cache.AddMessage(ctx, msg); err != nil {
			log.Printf("Failed to add message to cache for room %s: %v", msg.RoomID, err)
		}
	}()
	c.Hub.BroadcastToRoom(msg.RoomID, msg)
	if err := c.Hub.Chat.StoreChatMessage(ctx, msg); err != nil {
		log.Printf("Failed to store message from user %s in room %s: %v", c.UserID, msg.RoomID, err)
	}
	go c.updateUnreadCount(ctx, msg.RoomID)
}

func (c *Client) updateUnreadCount(ctx context.Context, roomID string) {
	// 这里可以实现更新未读消息数的逻辑，例如调用缓存服务或数据库
	var err error
	var users []string
	if users, err = c.Hub.Chat.GetChatRoomUsers(ctx, roomID); err != nil {
		// 更新未读消息数的逻辑
		log.Printf("Updating unread count for room %s, users: %v", roomID, users)
		return
	}
	for _, userID := range users {
		if userID != c.UserID {
			log.Printf("Incrementing unread count for user %s in room %s", userID, roomID)
			// 这里可以调用缓存服务或数据库来增加未读消息数
			if err = c.Hub.Cache.IncrUnread(ctx, userID, roomID); err != nil {
				log.Printf("Failed to increment unread count for user %s in room %s: %v", userID, roomID, err)
			}
		}

	}
}

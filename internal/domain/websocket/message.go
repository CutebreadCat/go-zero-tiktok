package websocket

import (
	"context"
	"log"
	"time"

	"go_zero-tiktok/internal/infra/cache"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

var pongWait = 60 * time.Second

// MessageCache 是消息缓存的依赖抽象
type MessageCache interface {
	AddMessage(ctx context.Context, message *types.MessageChat) (string, error)
	IncrUnread(ctx context.Context, userID, roomID string) error
	GetUnreadMessages(ctx context.Context, userID, roomID string, count int64) ([]cache.CacheMessage, error)
	GetUnreadCount(ctx context.Context, userID, roomID string) (int64, error)
	ClearUnread(ctx context.Context, userID, roomID string) error
}

// MessageRepository 是消息持久化的依赖抽象
type MessageRepository interface {
	StoreChatMessage(ctx context.Context, message *types.MessageChat) error
}

// MessageManager 管理消息的处理和分发
type MessageManager interface {
	HandleMessage(ctx context.Context, client *Client, msg *types.MessageChat)
	HandleGetUnread(ctx context.Context, client *Client, roomID string)
}

// messageManager 是 MessageManager 的实现
type messageManager struct {
	cache    MessageCache
	repo     MessageRepository
	roomRepo RoomRepository
	rooms    RoomManager
}

func (mm *messageManager) HandleMessage(ctx context.Context, client *Client, msg *types.MessageChat) {
	if !mm.rooms.IsMember(client, msg.RoomID) {
		log.Printf("User %s is not a member of room %s, ignoring message", client.UserID, msg.RoomID)
		return
	}

	go func() {
		if _, err := mm.cache.AddMessage(ctx, msg); err != nil {
			log.Printf("Failed to add message to cache for room %s: %v", msg.RoomID, err)
		}
	}()

	// 广播消息，包装成带 typek 的格式
	broadcastMsg := Message{
		Message: *msg,
		Typek:   "message",
	}
	mm.rooms.BroadcastToRoom(msg.RoomID, broadcastMsg)

	if err := mm.repo.StoreChatMessage(ctx, msg); err != nil {
		log.Printf("Failed to store message from user %s in room %s: %v", client.UserID, msg.RoomID, err)
	}
	go mm.UpdateUnreadCount(ctx, client.UserID, msg.RoomID)
}

// UnreadResponse 未读消息响应
type UnreadResponse struct {
	Typek    string               `json:"typek"`
	RoomID   string               `json:"room_id"`
	Count    int64                `json:"count"`
	Messages []cache.CacheMessage `json:"messages"`
}

func (mm *messageManager) HandleGetUnread(ctx context.Context, client *Client, roomID string) {
	if !mm.rooms.IsMember(client, roomID) {
		log.Printf("User %s is not a member of room %s, ignoring get_unread", client.UserID, roomID)
		return
	}

	// 获取未读消息数量
	count, err := mm.cache.GetUnreadCount(ctx, client.UserID, roomID)
	if err != nil {
		log.Printf("Failed to get unread count for user %s in room %s: %v", client.UserID, roomID, err)
		return
	}

	// 获取未读消息列表
	var messages []cache.CacheMessage
	if count > 0 {
		messages, err = mm.cache.GetUnreadMessages(ctx, client.UserID, roomID, count)
		if err != nil {
			log.Printf("Failed to get unread messages for user %s in room %s: %v", client.UserID, roomID, err)
			return
		}
	}

	// 发送给客户端
	resp := UnreadResponse{
		Typek:    "unread_messages",
		RoomID:   roomID,
		Count:    count,
		Messages: messages,
	}
	client.Send <- resp

	// 清除未读计数
	if err := mm.cache.ClearUnread(ctx, client.UserID, roomID); err != nil {
		log.Printf("Failed to clear unread for user %s in room %s: %v", client.UserID, roomID, err)
	}
}

func (mm *messageManager) UpdateUnreadCount(ctx context.Context, senderID, roomID string) {
	users, err := mm.roomRepo.GetChatRoomUsers(ctx, roomID)
	if err != nil {
		log.Printf("Failed to get room users for room %s: %v", roomID, err)
		return
	}
	log.Printf("Updating unread count for room %s, users: %v", roomID, users)
	for _, userID := range users {
		if userID != senderID {
			if err = mm.cache.IncrUnread(ctx, userID, roomID); err != nil {
				log.Printf("Failed to increment unread count for user %s in room %s: %v", userID, roomID, err)
			}
		}
	}
}

// ==================== Client 读写循环 ====================

func (c *Client) ReadLoop() {
	ctx, hand := context.WithTimeout(context.Background(), time.Hour*24)
	defer func() {
		c.Hub.Presence().RemoveClient(ctx, c)
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
		msg.Message.ID = uuid.New().String()
		msg.Message.SenderID = c.UserID
		msg.Message.CreatedAt = myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05")
		switch msg.Typek {
		case "message":
			c.Hub.Messages().HandleMessage(ctx, c, &msg.Message)
		case "get_unread":
			c.Hub.Messages().HandleGetUnread(ctx, c, msg.Message.RoomID)
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

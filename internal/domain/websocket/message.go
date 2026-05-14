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

const (
	MessageTopic = "chat-messages"
)

// ==================== MessageManager 实现 ====================

type messageManager struct {
	cache    MessageCache
	repo     MessageRepository
	roomRepo RoomRepository
	rooms    RoomManager
	writer   MessageWriter
}

func NewMessageManager(cache MessageCache, repo MessageRepository, roomRepo RoomRepository, rooms RoomManager, writer MessageWriter) MessageManager {
	return &messageManager{
		cache:    cache,
		repo:     repo,
		roomRepo: roomRepo,
		rooms:    rooms,
		writer:   writer,
	}
}

func (mm *messageManager) HandleMessage(ctx context.Context, client *Client, msg *types.MessageChat) {
	if !mm.rooms.IsMember(client, msg.RoomID) {
		log.Printf("User %s is not a member of room %s, ignoring message", client.UserID, msg.RoomID)
		return
	}

	// 1. 立即广播（实时）
	broadcastMsg := Message{
		Message: *msg,
		Typek:   "message",
	}
	mm.rooms.BroadcastToRoom(msg.RoomID, broadcastMsg)

	// 2. 异步：发送到 MQ 做持久化 + 更新未读
	event := NewMessageEvent(MessageTopic, msg.RoomID, client.UserID, msg.Content)
	log.Printf("Sending message event to Kafka: %+v", event)
	if err := mm.writer.SendMessage(ctx, event); err != nil {
		log.Printf("Failed to send message event for user %s in room %s: %v", client.UserID, msg.RoomID, err)
	} else {
		log.Printf("Message event sent to Kafka for user %s in room %s", client.UserID, msg.RoomID)
	}
}

func (mm *messageManager) HandleGetUnread(ctx context.Context, client *Client, roomID string) {
	if !mm.rooms.IsMember(client, roomID) {
		log.Printf("User %s is not a member of room %s, ignoring get_unread", client.UserID, roomID)
		return
	}
	count, err := mm.cache.GetUnreadCount(ctx, client.UserID, roomID)
	if err != nil {
		log.Printf("Failed to get unread count for user %s in room %s: %v", client.UserID, roomID, err)
		return
	}

	var messages []cache.CacheMessage
	if count > 0 {
		messages, err = mm.cache.GetUnreadMessages(ctx, client.UserID, roomID, count)
		if err != nil {
			log.Printf("Failed to get unread messages for user %s in room %s: %v", client.UserID, roomID, err)
			return
		}
	}
	for _, message := range messages {
		client.Send <- message
	}
	// 发送 MQ 事件，让消费端异步处理
	event := NewUnreadEvent(MessageTopic, roomID, client.UserID)
	if err := mm.writer.SendMessage(ctx, event); err != nil {
		log.Printf("Failed to send unread event for user %s in room %s: %v", client.UserID, roomID, err)
	}
}

func (mm *messageManager) HandleMessageByUserID(ctx context.Context, userID string, msg *types.MessageChat) {
	log.Printf("HandleMessageByUserID called, userID: %s, msg: %+v", userID, msg)

	// 生成消息 ID（如果为空）
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// 1. 添加到缓存
	if _, err := mm.cache.AddMessage(ctx, msg); err != nil {
		log.Printf("Failed to add message to cache for room %s: %v", msg.RoomID, err)
	} else {
		log.Printf("Message added to cache for room %s", msg.RoomID)
	}

	// 2. 持久化到数据库
	if err := mm.repo.StoreChatMessage(ctx, msg); err != nil {
		log.Printf("Failed to store message from user %s in room %s: %v", userID, msg.RoomID, err)
	} else {
		log.Printf("Message stored to database for room %s", msg.RoomID)
	}

	// 3. 更新未读计数
	mm.UpdateUnreadCount(ctx, userID, msg.RoomID)
}

func (mm *messageManager) HandleGetUnreadByUserID(ctx context.Context, userID, roomID string) {

	log.Printf("未读消息: user=%s, room=%s", userID, roomID)

	// 清除未读计数
	if err := mm.cache.ClearUnread(ctx, userID, roomID); err != nil {
		log.Printf("Failed to clear unread for user %s in room %s: %v", userID, roomID, err)
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

func (mm *messageManager) SendUnreadToUser(ctx context.Context, user *Client, messages []cache.CacheMessage) bool {
	for _, message := range messages {
		user.Send <- message
	}
	return false
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

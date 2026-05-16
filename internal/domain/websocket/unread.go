package websocket

import (
	"context"
	"log"
)

// ==================== 未读消息处理 ====================

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

	var messages []CacheMessage
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

func (mm *messageManager) SendUnreadToUser(ctx context.Context, user *Client, messages []CacheMessage) bool {
	for _, message := range messages {
		user.Send <- message
	}
	return false
}

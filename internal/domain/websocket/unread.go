package websocket

import (
	"context"
	"log"
)

func (mm *messageManager) HandleGetUnread(ctx context.Context, client *Client, roomID string) {
	if !mm.rooms.IsMember(client, roomID) {
		log.Printf("用户 %s 不是房间 %s 的成员，忽略获取未读", client.UserID, roomID)
		return
	}
	count, err := mm.cache.GetUnreadCount(ctx, client.UserID, roomID)
	if err != nil {
		log.Printf("获取用户 %s 在房间 %s 的未读计数失败: %v", client.UserID, roomID, err)
		return
	}

	var messages []CacheMessage
	if count > 0 {
		messages, err = mm.cache.GetMessages(ctx, client.UserID, roomID, count)
		if err != nil {
			log.Printf("获取用户 %s 在房间 %s 的未读消息失败: %v", client.UserID, roomID, err)
			return
		}
	}
	for _, message := range messages {
		client.Send <- message
	}
	event := NewUnreadEvent(MessageTopic, roomID, client.UserID)
	if err := mm.writer.SendMessage(ctx, event); err != nil {
		log.Printf("发送未读事件到 Kafka 失败 (用户 %s, 房间 %s): %v", client.UserID, roomID, err)
	}
}

func (mm *messageManager) HandleGetUnreadByUserID(ctx context.Context, userID, roomID string) {

	log.Printf("处理未读消息: user=%s, room=%s", userID, roomID)

	if err := mm.cache.ClearUnread(ctx, userID, roomID); err != nil {
		log.Printf("清除用户 %s 在房间 %s 的未读失败: %v", userID, roomID, err)
	}
}

func (mm *messageManager) UpdateUnreadCount(ctx context.Context, senderID, roomID string) {
	users, err := mm.roomRepo.GetChatRoomUsers(ctx, roomID)
	if err != nil {
		log.Printf("获取房间 %s 的用户列表失败: %v", roomID, err)
		return
	}
	log.Printf("更新房间 %s 的未读计数，用户: %v", roomID, users)
	for _, userID := range users {
		if userID != senderID {
			if mm.rooms.IsOnline(userID) {
				log.Printf("用户 %s 在线，跳过未读计数增加", userID)
				continue
			}
			if err = mm.cache.IncrUnread(ctx, userID, roomID); err != nil {
				log.Printf("增加用户 %s 在房间 %s 的未读计数失败: %v", userID, roomID, err)
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

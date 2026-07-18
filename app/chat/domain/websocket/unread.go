package websocket

import (
	"context"
	appLogger "go_zero-tiktok/Prometheus/logger"
)

func (mm *messageManager) HandleGetUnread(ctx context.Context, client *Client, roomID string) {
	if !mm.rooms.IsMember(client, roomID) {
		appLogger.Infof("鐢ㄦ埛 %s 涓嶆槸鎴块棿 %s 鐨勬垚鍛橈紝蹇界暐鑾峰彇鏈", client.UserID, roomID)
		return
	}
	count, err := mm.cache.GetUnreadCount(ctx, client.UserID, roomID)
	if err != nil {
		appLogger.Infof("鑾峰彇鐢ㄦ埛 %s 鍦ㄦ埧闂?%s 鐨勬湭璇昏鏁板け璐? %v", client.UserID, roomID, err)
		return
	}

	var messages []CacheMessage
	if count > 0 {
		messages, err = mm.cache.GetMessages(ctx, client.UserID, roomID, count)
		if err != nil {
			appLogger.Infof("鑾峰彇鐢ㄦ埛 %s 鍦ㄦ埧闂?%s 鐨勬湭璇绘秷鎭け璐? %v", client.UserID, roomID, err)
			return
		}
	}
	for _, message := range messages {
		client.Send <- message
	}
	event := NewUnreadEvent(MessageTopic, roomID, client.UserID)
	if err := mm.writer.SendMessage(ctx, event); err != nil {
		appLogger.Infof("鍙戦€佹湭璇讳簨浠跺埌 Kafka 澶辫触 (鐢ㄦ埛 %s, 鎴块棿 %s): %v", client.UserID, roomID, err)
	}
}

func (mm *messageManager) HandleGetUnreadByUserID(ctx context.Context, userID, roomID string) {

	appLogger.Infof("澶勭悊鏈娑堟伅: user=%s, room=%s", userID, roomID)

	if err := mm.cache.ClearUnread(ctx, userID, roomID); err != nil {
		appLogger.Infof("娓呴櫎鐢ㄦ埛 %s 鍦ㄦ埧闂?%s 鐨勬湭璇诲け璐? %v", userID, roomID, err)
	}
}

func (mm *messageManager) UpdateUnreadCount(ctx context.Context, senderID, roomID string) {
	users, err := mm.roomRepo.GetChatRoomUsers(ctx, roomID)
	if err != nil {
		appLogger.Infof("鑾峰彇鎴块棿 %s 鐨勭敤鎴峰垪琛ㄥけ璐? %v", roomID, err)
		return
	}
	appLogger.Infof("鏇存柊鎴块棿 %s 鐨勬湭璇昏鏁帮紝鐢ㄦ埛: %v", roomID, users)
	for _, userID := range users {
		if userID != senderID {
			if mm.rooms.IsOnline(userID) {
				appLogger.Infof("用户 %s 在线，跳过未读计数增加", userID)
				continue
			}
			if err = mm.cache.IncrUnread(ctx, userID, roomID); err != nil {
				appLogger.Infof("澧炲姞鐢ㄦ埛 %s 鍦ㄦ埧闂?%s 鐨勬湭璇昏鏁板け璐? %v", userID, roomID, err)
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

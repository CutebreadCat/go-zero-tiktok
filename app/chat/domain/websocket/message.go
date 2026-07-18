package websocket

import (
	"context"
	appLogger "go_zero-tiktok/Prometheus/logger"

	"go_zero-tiktok/pkg/contract"

	"github.com/google/uuid"
)

type messageManager struct {
	cache    MessageCache
	repo     MessageRepository
	roomRepo RoomRepository
	rooms    RoomManager
	writer   MessageWriter
	ai       *AIChat
}

func NewMessageManager(cache MessageCache, repo MessageRepository, roomRepo RoomRepository, rooms RoomManager, writer MessageWriter, ai *AIChat) MessageManager {
	return &messageManager{
		cache:    cache,
		repo:     repo,
		roomRepo: roomRepo,
		rooms:    rooms,
		writer:   writer,
		ai:       ai,
	}
}

func (mm *messageManager) HandleMessage(ctx context.Context, client *Client, msg *types.MessageChat) {
	if !mm.rooms.IsMember(client, msg.RoomID) {
		appLogger.Infof("鐢ㄦ埛 %s 涓嶆槸鎴块棿 %s 鐨勬垚鍛橈紝蹇界暐娑堟伅", client.UserID, msg.RoomID)
		return
	}

	broadcastMsg := Message{
		Message: *msg,
		Typek:   EventTypeMessage,
	}
	mm.rooms.BroadcastToRoom(msg.RoomID, broadcastMsg)

	event := NewMessageEvent(MessageTopic, msg.RoomID, client.UserID, msg.Content)
	if err := mm.writer.SendMessage(ctx, event); err != nil {
		appLogger.Infof("鍙戦€佹秷鎭簨浠跺埌 Kafka 澶辫触 (鐢ㄦ埛 %s, 鎴块棿 %s): %v", client.UserID, msg.RoomID, err)
	}

	go func(userID, roomID, content string) {
		reached, err := mm.ai.CheckAndEnqueue(ctx, userID, roomID, msg)
		if err != nil {
			appLogger.Infof("妫€鏌?AI 鍏ラ槦澶辫触 (鐢ㄦ埛 %s, 鎴块棿 %s): %v", userID, roomID, err)
			return
		}
		if reached {
			aiEvent := NewAIChatEvent(MessageTopic, roomID, userID, content)
			if err := mm.writer.SendMessage(ctx, aiEvent); err != nil {
				appLogger.Infof("鍙戦€?AI 鑱婂ぉ浜嬩欢鍒?Kafka 澶辫触 (鐢ㄦ埛 %s, 鎴块棿 %s): %v", userID, roomID, err)
			}
		}
	}(client.UserID, msg.RoomID, msg.Content)
}

func (mm *messageManager) HandleMessageByUserID(ctx context.Context, userID string, msg *types.MessageChat) {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	if _, err := mm.cache.AddMessage(ctx, msg); err != nil {
		appLogger.Infof("娣诲姞娑堟伅鍒扮紦瀛樺け璐?(鎴块棿 %s): %v", msg.RoomID, err)
	}

	if err := mm.repo.StoreChatMessage(ctx, msg); err != nil {
		appLogger.Infof("瀛樺偍娑堟伅鍒版暟鎹簱澶辫触 (鐢ㄦ埛 %s, 鎴块棿 %s): %v", userID, msg.RoomID, err)
	}

	mm.UpdateUnreadCount(ctx, userID, msg.RoomID)

}

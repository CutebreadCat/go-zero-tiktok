package chat

import (
	"context"
	"go_zero-tiktok/internal/svc/xerr"
	"log"

	"gorm.io/gorm"
)

func CreateChatRoom(ctx context.Context, db *gorm.DB, chat *UserChat) error {
	if err := db.Model(chat).Create(chat).Error; err != nil {
		log.Printf("CreateChatRoom error: %v", err)
		return xerr.New(500, "创建聊天室失败")
	}
	return nil
}

func CreateChatMessage(ctx context.Context, db *gorm.DB, message *MessageChat) error {
	if err := db.Model(message).Create(message).Error; err != nil {
		log.Printf("CreateChatMessage error: %v", err)
		return xerr.New(500, "创建消息失败")
	}
	return nil
}

func GetChatRoomMessage(ctx context.Context, db *gorm.DB, room_id string, page_size int, page_num int) ([]*MessageChat, error) {
	var messages []*MessageChat
	if err := db.Model(&MessageChat{}).Where("room_id = ?", room_id).Find(&messages).Scopes(func(db *gorm.DB) *gorm.DB {
		return db.Limit(page_size).Offset((page_num - 1) * page_size)
	}).Error; err != nil {

		return nil, xerr.New(404, "聊天室消息未找到")
	}
	return messages, nil
}

func GetJoinRooms(ctx context.Context, db *gorm.DB, user_id string) ([]string, error) {
	var roomsid []string
	if err := db.Model(&UserChat{}).Where("user_id = ?", user_id).Pluck("room_id", &roomsid).Error; err != nil {
		log.Printf("GetJoinRooms error: %v", err)
		return nil, xerr.New(404, "用户加入的聊天室未找到")
	}
	return roomsid, nil
}

func JoinChatRoom(ctx context.Context, db *gorm.DB, user_id string, room_id string) error {
	var existingChat UserChat
	if err := db.Model(&UserChat{}).Where("user_id = ? AND room_id = ?", user_id, room_id).First(&existingChat).Error; err == nil {
		return xerr.New(400, "用户已加入该聊天室")
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("JoinChatRoom error: %v", err)
		return xerr.New(500, "查询加入聊天室失败")
	}
	var roomInfo UserChat
	if err := db.Model(&UserChat{}).Where("room_id = ?", room_id).First(&roomInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.New(404, "聊天室不存在")
		}
		log.Printf("JoinChatRoom error: %v", err)
		return xerr.New(500, "查询聊天室信息失败")
	}
	if roomInfo.Leix == 0 {
		return xerr.New(400, "这是私人聊天室,你不能加入")
	}
	chat := &UserChat{
		UserID:   user_id,
		RoomID:   room_id,
		Leix:     roomInfo.Leix,
		RoomName: roomInfo.RoomName,
	}
	if err := db.Model(chat).Create(chat).Error; err != nil {
		log.Printf("JoinChatRoom error: %v", err)
		return xerr.New(500, "加入聊天室失败")
	}
	return nil
}

func StoreChatMessage(ctx context.Context, db *gorm.DB, message *MessageChat) error {
	if err := db.Model(message).Create(message).Error; err != nil {
		log.Printf("StoreChatMessage error: %v", err)
		return xerr.New(500, "存储消息失败")
	}
	return nil
}
func GetChatRoomUsers(ctx context.Context, db *gorm.DB, room_id string) ([]string, error) {
	var usersid []string
	if err := db.Model(&UserChat{}).Where("room_id = ?", room_id).Pluck("user_id", &usersid).Error; err != nil {
		log.Printf("GetChatRoomUsers error: %v", err)
		return nil, xerr.New(404, "聊天室用户未找到")
	}
	return usersid, nil
}

package chat

import (
	"context"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	"log"

	"gorm.io/gorm"
)

func CreateChatRoom(ctx context.Context, db *gorm.DB, chat *types.User_chat) error {
	if err := db.Model(chat).Create(chat).Error; err != nil {
		log.Printf("CreateChatRoom error: %v", err)
		return xerr.New(500, "创建聊天室失败")
	}
	return nil
}

func GetChatRoomMessage(ctx context.Context, db *gorm.DB, room_id string, page_size int, page_num int) ([]*types.MessageChat, error) {
	var messages []*types.MessageChat
	if err := db.Model(&types.MessageChat{}).Where("room_id = ?", room_id).Find(&messages).Scopes(func(db *gorm.DB) *gorm.DB {
		return db.Limit(page_size).Offset((page_num - 1) * page_size)
	}).Error; err != nil {

		return nil, xerr.New(404, "聊天室消息未找到")
	}
	return messages, nil
}

func GetJoinRooms(ctx context.Context, db *gorm.DB, user_id string) ([]string, error) {
	var roomsid []string
	if err := db.Model(&types.User_chat{}).Where("user_id = ?", user_id).Pluck("room_id", &roomsid).Error; err != nil {
		log.Printf("GetJoinRooms error: %v", err)
		return nil, xerr.New(404, "用户加入的聊天室未找到")
	}
	return roomsid, nil
}

func JoinChatRoom(ctx context.Context, db *gorm.DB, user_id string, room_id string) error {
	var existingChat types.User_chat
	if err := db.Model(&types.User_chat{}).Where("user_id = ? AND room_id = ?", user_id, room_id).First(&existingChat).Error; err == nil {
		return xerr.New(400, "用户已加入该聊天室")
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("JoinChatRoom error: %v", err)
		return xerr.New(500, "查询加入聊天室失败")
	}
	if existingChat.Leix == 0 {
		return xerr.New(400, "这是私人聊天室,你不能加入")
	}
	chat := &types.User_chat{
		UserID: user_id,
		RoomID: room_id,
	}
	if err := db.Model(chat).Create(chat).Error; err != nil {
		log.Printf("JoinChatRoom error: %v", err)
		return xerr.New(500, "加入聊天室失败")
	}
	return nil
}

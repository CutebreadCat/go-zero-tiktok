package chat

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/svc/xerr"

	"gorm.io/gorm"
)

func CreateChatRoom(ctx context.Context, db *gorm.DB, chat *UserChat) error {
	if err := db.Model(chat).Create(chat).Error; err != nil {
		return xerr.Wrap(err, "create chat room failed")
	}
	return nil
}

func CreateChatMessage(ctx context.Context, db *gorm.DB, message *MessageChat) error {
	if err := db.Model(message).Create(message).Error; err != nil {
		return xerr.Wrap(err, "create chat message failed")
	}
	return nil
}

func GetChatRoomMessage(ctx context.Context, db *gorm.DB, room_id string, page_size int, page_num int) ([]*MessageChat, error) {
	var messages []*MessageChat
	if err := db.Model(&MessageChat{}).Where("room_id = ?", room_id).Find(&messages).Scopes(func(db *gorm.DB) *gorm.DB {
		return db.Limit(page_size).Offset((page_num - 1) * page_size)
	}).Error; err != nil {
		return nil, xerr.Wrap(err, "get chat room message failed")
	}
	return messages, nil
}

func GetJoinRooms(ctx context.Context, db *gorm.DB, user_id string) ([]string, error) {
	var roomsid []string
	if err := db.Model(&UserChat{}).Where("user_id = ?", user_id).Pluck("room_id", &roomsid).Error; err != nil {
		return nil, xerr.Wrap(err, "get join rooms failed")
	}
	return roomsid, nil
}

func JoinChatRoom(ctx context.Context, db *gorm.DB, user_id string, room_id string) error {
	var existingChat UserChat
	if err := db.Model(&UserChat{}).Where("user_id = ? AND room_id = ?", user_id, room_id).First(&existingChat).Error; err == nil {
		return xerr.NewInvalidParam("用户已加入该聊天室")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.Wrap(err, "query join chat room failed")
	}
	var roomInfo UserChat
	if err := db.Model(&UserChat{}).Where("room_id = ?", room_id).First(&roomInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xerr.NewInvalidParam("聊天室不存在")
		}
		return xerr.Wrap(err, "query chat room info failed")
	}
	if roomInfo.Leix == 0 {
		return xerr.NewInvalidParam("这是私人聊天室,你不能加入")
	}
	chat := &UserChat{
		UserID:   user_id,
		RoomID:   room_id,
		Leix:     roomInfo.Leix,
		RoomName: roomInfo.RoomName,
	}
	if err := db.Model(chat).Create(chat).Error; err != nil {
		return xerr.Wrap(err, "join chat room failed")
	}
	return nil
}

func StoreChatMessage(ctx context.Context, db *gorm.DB, message *MessageChat) error {
	if err := db.Model(message).Create(message).Error; err != nil {
		return xerr.Wrap(err, "store chat message failed")
	}
	return nil
}

func GetChatRoomUsers(ctx context.Context, db *gorm.DB, room_id string) ([]string, error) {
	var usersid []string
	if err := db.Model(&UserChat{}).Where("room_id = ?", room_id).Pluck("user_id", &usersid).Error; err != nil {
		return nil, xerr.Wrap(err, "get chat room users failed")
	}
	return usersid, nil
}

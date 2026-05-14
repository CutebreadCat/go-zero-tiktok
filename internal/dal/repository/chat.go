package repository

import (
	"context"

	chattable "go_zero-tiktok/internal/dal/tables/chat"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"gorm.io/gorm"
)

// UserChatToResponse 将数据库模型转换为API响应类型
func (r *ChatRepo) UserChatToResponse(chat *chattable.UserChat) types.User_chat {
	return types.User_chat{
		UserID:   chat.UserID,
		RoomID:   chat.RoomID,
		Leix:     chat.Leix,
		RoomName: chat.RoomName,
	}
}

// MessageToResponse 将数据库模型转换为API响应类型
func (r *ChatRepo) MessageToResponse(msg *chattable.MessageChat) types.MessageChat {
	return types.MessageChat{
		ID:        msg.ID,
		RoomID:    msg.RoomID,
		SenderID:  msg.SenderID,
		Content:   msg.Content,
		CreatedAt: myutils.TimeToStr(msg.CreatedAt, ""),
	}
}

// MessagesToResponse 将数据库模型切片转换为API响应类型切片
func (r *ChatRepo) MessagesToResponse(msgs []*chattable.MessageChat) []types.MessageChat {
	result := make([]types.MessageChat, 0, len(msgs))
	for _, msg := range msgs {
		if msg != nil {
			result = append(result, r.MessageToResponse(msg))
		}
	}
	return result
}

type ChatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) CreateChatRoom(ctx context.Context, chat *chattable.UserChat) error {
	return chattable.CreateChatRoom(ctx, r.db, chat)
}

// CreateChatRoomFromParams 通过参数创建聊天室，logic层不需要知道数据库模型
func (r *ChatRepo) CreateChatRoomFromParams(ctx context.Context, userID, roomID string, leix int32, roomName string) error {
	chat := &chattable.UserChat{
		UserID:   userID,
		RoomID:   roomID,
		Leix:     leix,
		RoomName: roomName,
	}
	return chattable.CreateChatRoom(ctx, r.db, chat)
}

func (r *ChatRepo) CreateChatMessage(ctx context.Context, message *types.MessageChat) error {
	// 将 API 类型转换为数据库模型
	createdAt, err := myutils.StrToTime(message.CreatedAt, "")
	if err != nil {
		return err
	}
	chatMsg := &chattable.MessageChat{
		ID:        message.ID,
		RoomID:    message.RoomID,
		SenderID:  message.SenderID,
		Content:   message.Content,
		CreatedAt: createdAt,
	}
	return chattable.CreateChatMessage(ctx, r.db, chatMsg)
}

func (r *ChatRepo) GetChatRoomMessage(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*chattable.MessageChat, error) {
	return chattable.GetChatRoomMessage(ctx, r.db, roomID, pageSize, pageNum)
}

func (r *ChatRepo) GetJoinRooms(ctx context.Context, userID string) ([]string, error) {
	return chattable.GetJoinRooms(ctx, r.db, userID)
}

func (r *ChatRepo) JoinChatRoom(ctx context.Context, userID string, roomID string) error {
	return chattable.JoinChatRoom(ctx, r.db, userID, roomID)
}
func (r *ChatRepo) StoreChatMessage(ctx context.Context, message *types.MessageChat) error {
	// 将 API 类型转换为数据库模型
	createdAt, err := myutils.StrToTime(message.CreatedAt, "")
	if err != nil {
		return err
	}
	chatMsg := &chattable.MessageChat{
		ID:        message.ID,
		RoomID:    message.RoomID,
		SenderID:  message.SenderID,
		Content:   message.Content,
		CreatedAt: createdAt,
	}
	return chattable.StoreChatMessage(ctx, r.db, chatMsg)
}
func (r *ChatRepo) GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	return chattable.GetChatRoomUsers(ctx, r.db, roomID)
}

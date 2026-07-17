package domain

import (
	"context"

	"go_zero-tiktok/pkg/contract"
)

type ChatService struct {
	chatRepo IChatRepo
}

func NewChatService(chatRepo IChatRepo) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
	}
}

// CreateChatRoom 创建聊天室
func (s *ChatService) CreateChatRoom(ctx context.Context, roomID string, roomType int32, roomName string, userIDs []string) error {
	for _, uid := range userIDs {
		if uid == "" {
			continue
		}
		if err := s.chatRepo.CreateChatRoomFromParams(ctx, uid, roomID, roomType, roomName); err != nil {
			return err
		}
	}
	return nil
}

// JoinChatRoom 加入聊天室
func (s *ChatService) JoinChatRoom(ctx context.Context, userID, roomID string) error {
	return s.chatRepo.JoinChatRoom(ctx, userID, roomID)
}

// GetJoinRooms 获取用户加入的房间列表
func (s *ChatService) GetJoinRooms(ctx context.Context, userID string) ([]string, error) {
	return s.chatRepo.GetJoinRooms(ctx, userID)
}

// GetMessages 获取房间消息列表
func (s *ChatService) GetMessages(ctx context.Context, roomID string, pageSize, pageNum int) ([]*types.MessageChat, error) {
	return s.chatRepo.GetChatRoomMessage(ctx, roomID, pageSize, pageNum)
}

// GetChatRoomUsers 获取房间用户列表
func (s *ChatService) GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	return s.chatRepo.GetChatRoomUsers(ctx, roomID)
}

// StoreChatMessage 存储聊天消息
func (s *ChatService) StoreChatMessage(ctx context.Context, message *types.MessageChat) error {
	return s.chatRepo.StoreChatMessage(ctx, message)
}

package repository

import (
	"context"

	chattable "go_zero-tiktok/internal/dal/tables/chat"
	"go_zero-tiktok/internal/types"

	"gorm.io/gorm"
)

type ChatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) CreateChatRoom(ctx context.Context, chat *types.User_chat) error {
	return chattable.CreateChatRoom(ctx, r.db, chat)
}

func (r *ChatRepo) CreateChatMessage(ctx context.Context, message *types.MessageChat) error {
	return chattable.CreateChatMessage(ctx, r.db, message)
}

func (r *ChatRepo) GetChatRoomMessage(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*types.MessageChat, error) {
	return chattable.GetChatRoomMessage(ctx, r.db, roomID, pageSize, pageNum)
}

func (r *ChatRepo) GetJoinRooms(ctx context.Context, userID string) ([]string, error) {
	return chattable.GetJoinRooms(ctx, r.db, userID)
}

func (r *ChatRepo) JoinChatRoom(ctx context.Context, userID string, roomID string) error {
	return chattable.JoinChatRoom(ctx, r.db, userID, roomID)
}
func (r *ChatRepo) StoreChatMessage(ctx context.Context, message *types.MessageChat) error {
	return chattable.StoreChatMessage(ctx, r.db, message)
}
func (r *ChatRepo) GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	return chattable.GetChatRoomUsers(ctx, r.db, roomID)
}

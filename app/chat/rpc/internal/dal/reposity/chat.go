package reposity

import (
	"context"

	chattable "go_zero-tiktok/app/chat/rpc/internal/dal/tables/chat"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

type ChatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) CreateChatRoomFromParams(ctx context.Context, userID, roomID string, leix int32, roomName string) error {
	chat := &chattable.UserChat{
		UserID:   userID,
		RoomID:   roomID,
		Leix:     leix,
		RoomName: roomName,
	}
	if err := chattable.CreateChatRoom(ctx, r.db, chat); err != nil {
		return pkgerrors.WithMessage(err, "ChatRepo.CreateChatRoomFromParams")
	}
	return nil
}

func (r *ChatRepo) MessageToResponse(msg *chattable.MessageChat) types.MessageChat {
	return types.MessageChat{
		ID:        msg.ID,
		RoomID:    msg.RoomID,
		SenderID:  msg.SenderID,
		Content:   msg.Content,
		CreatedAt: myutils.TimeToStr(msg.CreatedAt, ""),
	}
}

func (r *ChatRepo) GetChatRoomMessage(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*types.MessageChat, error) {
	msgs, err := chattable.GetChatRoomMessage(ctx, r.db, roomID, pageSize, pageNum)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "ChatRepo.GetChatRoomMessage")
	}
	result := make([]*types.MessageChat, 0, len(msgs))
	for _, msg := range msgs {
		if msg != nil {
			resp := r.MessageToResponse(msg)
			result = append(result, &resp)
		}
	}
	return result, nil
}

func (r *ChatRepo) GetJoinRooms(ctx context.Context, userID string) ([]string, error) {
	rooms, err := chattable.GetJoinRooms(ctx, r.db, userID)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "ChatRepo.GetJoinRooms")
	}
	return rooms, nil
}

func (r *ChatRepo) JoinChatRoom(ctx context.Context, userID string, roomID string) error {
	if err := chattable.JoinChatRoom(ctx, r.db, userID, roomID); err != nil {
		return pkgerrors.WithMessage(err, "ChatRepo.JoinChatRoom")
	}
	return nil
}

func (r *ChatRepo) GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	users, err := chattable.GetChatRoomUsers(ctx, r.db, roomID)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "ChatRepo.GetChatRoomUsers")
	}
	return users, nil
}

func (r *ChatRepo) StoreChatMessage(ctx context.Context, message *types.MessageChat) error {
	createdAt, err := myutils.StrToTime(message.CreatedAt, "")
	if err != nil {
		return pkgerrors.WithMessage(err, "ChatRepo.StoreChatMessage: parse created_at")
	}
	chatMsg := &chattable.MessageChat{
		ID:        message.ID,
		RoomID:    message.RoomID,
		SenderID:  message.SenderID,
		Content:   message.Content,
		CreatedAt: createdAt,
	}
	if err := chattable.StoreChatMessage(ctx, r.db, chatMsg); err != nil {
		return pkgerrors.WithMessage(err, "ChatRepo.StoreChatMessage")
	}
	return nil
}

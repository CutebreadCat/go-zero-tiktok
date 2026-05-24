package chat

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestGetMessagesLogic_GetMessages(t *testing.T) {
	tests := []struct {
		name     string
		req      *types.GetMessagesRequest
		chatRepo *mock.ChatRepo
		wantErr  bool
		wantLen  int
	}{
		{
			name:    "empty room id",
			req:     &types.GetMessagesRequest{RoomID: ""},
			wantErr: true,
		},
		{
			name: "get messages failed",
			req:  &types.GetMessagesRequest{RoomID: "r1"},
			chatRepo: &mock.ChatRepo{
				GetChatRoomMessageFn: func(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*types.MessageChat, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success with messages",
			req:  &types.GetMessagesRequest{RoomID: "r1", PageSize: 10, PageNumber: 1},
			chatRepo: &mock.ChatRepo{
				GetChatRoomMessageFn: func(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*types.MessageChat, error) {
					return []*types.MessageChat{
						{ID: "m1", RoomID: "r1", SenderID: "u1", Content: "hello"},
						{ID: "m2", RoomID: "r1", SenderID: "u2", Content: "hi"},
					}, nil
				},
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "success empty messages",
			req:  &types.GetMessagesRequest{RoomID: "r1"},
			chatRepo: &mock.ChatRepo{
				GetChatRoomMessageFn: func(ctx context.Context, roomID string, pageSize int, pageNum int) ([]*types.MessageChat, error) {
					return nil, nil
				},
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, nil, nil, nil, tt.chatRepo)
			logic := NewGetMessagesLogic(context.Background(), svcCtx)
			resp, err := logic.GetMessages(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(resp.Messages) != tt.wantLen {
				t.Errorf("expected %d messages, got %d", tt.wantLen, len(resp.Messages))
			}
		})
	}
}

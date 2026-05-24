package chat

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestCreateChatRoomLogic_CreateChatRoom(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		req      *types.CreateChatRoomRequest
		chatRepo *mock.ChatRepo
		wantErr  bool
	}{
		{
			name:    "invalid type",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.CreateChatRoomRequest{Types: 2},
			wantErr: true,
		},
		{
			name:    "group chat empty room name",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.CreateChatRoomRequest{Types: 1, RoomName: ""},
			wantErr: true,
		},
		{
			name:    "private chat no user ids",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.CreateChatRoomRequest{Types: 0, UserIDs: []string{}},
			wantErr: true,
		},
		{
			name: "create room failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.CreateChatRoomRequest{Types: 0, UserIDs: []string{"u2"}},
			chatRepo: &mock.ChatRepo{
				CreateChatRoomFromParamsFn: func(ctx context.Context, userID, roomID string, leix int32, roomName string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success private chat",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.CreateChatRoomRequest{Types: 0, UserIDs: []string{"u2"}},
			chatRepo: &mock.ChatRepo{
				CreateChatRoomFromParamsFn: func(ctx context.Context, userID, roomID string, leix int32, roomName string) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "success group chat",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.CreateChatRoomRequest{Types: 1, RoomName: "test-group", UserIDs: []string{"u2", "u3"}},
			chatRepo: &mock.ChatRepo{
				CreateChatRoomFromParamsFn: func(ctx context.Context, userID, roomID string, leix int32, roomName string) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, nil, nil, nil, tt.chatRepo)
			logic := NewCreateChatRoomLogic(tt.ctx, svcCtx)
			resp, err := logic.CreateChatRoom(tt.req)
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
			if resp.RoomID == "" {
				t.Errorf("expected room ID, got empty")
			}
		})
	}
}

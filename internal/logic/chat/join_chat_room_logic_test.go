package chat

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestJoinChatRoomLogic_JoinChatRoom(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		req      *types.JoinChatRoomRequest
		chatRepo *mock.ChatRepo
		wantErr  bool
	}{
		{
			name:    "empty room id",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			req:     &types.JoinChatRoomRequest{RoomID: ""},
			wantErr: true,
		},
		{
			name: "join failed",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.JoinChatRoomRequest{RoomID: "r1"},
			chatRepo: &mock.ChatRepo{
				JoinChatRoomFn: func(ctx context.Context, userID string, roomID string) error {
					return errors.New("already joined")
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			req:  &types.JoinChatRoomRequest{RoomID: "r1"},
			chatRepo: &mock.ChatRepo{
				JoinChatRoomFn: func(ctx context.Context, userID string, roomID string) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, nil, nil, nil, tt.chatRepo)
			logic := NewJoinChatRoomLogic(tt.ctx, svcCtx)
			_, err := logic.JoinChatRoom(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

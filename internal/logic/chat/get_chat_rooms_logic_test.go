package chat

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/svc/mock"
	"go_zero-tiktok/internal/types"
)

func TestGetChatRoomsLogic_GetChatRooms(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		chatRepo *mock.ChatRepo
		wantErr  bool
		wantLen  int
	}{
		{
			name:    "get rooms failed",
			ctx:     context.WithValue(context.Background(), "user_id", "u1"),
			chatRepo: &mock.ChatRepo{
				GetJoinRoomsFn: func(ctx context.Context, userID string) ([]string, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "success with rooms",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			chatRepo: &mock.ChatRepo{
				GetJoinRoomsFn: func(ctx context.Context, userID string) ([]string, error) {
					return []string{"r1", "r2", "r3"}, nil
				},
			},
			wantErr: false,
			wantLen: 3,
		},
		{
			name: "success empty rooms",
			ctx:  context.WithValue(context.Background(), "user_id", "u1"),
			chatRepo: &mock.ChatRepo{
				GetJoinRoomsFn: func(ctx context.Context, userID string) ([]string, error) {
					return []string{}, nil
				},
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := mock.NewServiceContext(nil, nil, nil, nil, nil, nil, tt.chatRepo)
			logic := NewGetChatRoomsLogic(tt.ctx, svcCtx)
			resp, err := logic.GetChatRooms(&types.GetChatRoomsRequest{})
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
			if len(resp.RoomsId) != tt.wantLen {
				t.Errorf("expected %d rooms, got %d", tt.wantLen, len(resp.RoomsId))
			}
		})
	}
}

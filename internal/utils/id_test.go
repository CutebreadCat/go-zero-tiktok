package myutils

import (
	"context"
	"strings"
	"testing"

	"go_zero-tiktok/internal/shared/ctxkey"
)

func TestGenerateIDs(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		fn     func() string
	}{
		{name: "user", prefix: "u", fn: GenerateUserID},
		{name: "video", prefix: "v", fn: GenerateVideoID},
		{name: "comment", prefix: "c", fn: GenerateCommentID},
		{name: "room", prefix: "r", fn: GenerateRoomID},
		{name: "message", prefix: "m", fn: GenerateMessageID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.fn()
			if !strings.HasPrefix(id, tt.prefix) {
				t.Fatalf("id = %q, want prefix %q", id, tt.prefix)
			}
			if len(id) <= len(tt.prefix) {
				t.Fatalf("id = %q, want generated suffix", id)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr bool
	}{
		{name: "nil context", ctx: nil, wantErr: true},
		{name: "typed key", ctx: context.WithValue(context.Background(), ctxkey.UserID, "u1"), want: "u1"},
		{name: "string key user_id", ctx: context.WithValue(context.Background(), "user_id", "u2"), want: "u2"},
		{name: "string key userId", ctx: context.WithValue(context.Background(), "userId", "u3"), want: "u3"},
		{name: "empty user", ctx: context.WithValue(context.Background(), ctxkey.UserID, ""), wantErr: true},
		{name: "missing user", ctx: context.Background(), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserIDFromContext(tt.ctx)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("user id = %q, want %q", got, tt.want)
			}
		})
	}
}

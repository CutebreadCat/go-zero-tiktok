package websocket

import (
	"context"
	"errors"
	"testing"
)

func TestLoadRooms(t *testing.T) {
	tests := []struct {
		name           string
		repo           *mockRoomRepo
		wantRooms      []string // 期望 client.Rooms 包含的房间
		wantJoinCalls  int      // 期望 JoinRoom 调用次数
	}{
		{
			name:          "正常加载房间",
			repo:          &mockRoomRepo{rooms: []string{"r1", "r2"}},
			wantRooms:     []string{"r1", "r2"},
			wantJoinCalls: 2,
		},
		{
			name:          "Repo错误跳过加载",
			repo:          &mockRoomRepo{roomsErr: errors.New("db down")},
			wantRooms:     nil,
			wantJoinCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockRoomCache{}
			rm := newTestRoomManager(tt.repo, cache)
			client := newTestClient("u1")

			rm.LoadRooms(context.Background(), client)

			// 检查 client.Rooms
			if tt.wantRooms == nil {
				if len(client.Rooms) != 0 {
					t.Errorf("client.Rooms should be empty, got %v", client.Rooms)
				}
			} else {
				for _, r := range tt.wantRooms {
					if !client.Rooms[r] {
						t.Errorf("client.Rooms[%s] should be true", r)
					}
				}
			}

			// 检查 JoinRoom 调用
			cache.mu.Lock()
			if len(cache.joinedRooms) != tt.wantJoinCalls {
				t.Errorf("JoinRoom called %d times, want %d", len(cache.joinedRooms), tt.wantJoinCalls)
			}
			cache.mu.Unlock()

			// 检查 rm.rooms
			if tt.wantRooms != nil {
				rm.mu.RLock()
				for _, r := range tt.wantRooms {
					if _, ok := rm.rooms[r][client]; !ok {
						t.Errorf("rm.rooms[%s][client] should be true", r)
					}
				}
				rm.mu.RUnlock()
			}
		})
	}
}

func TestRemoveFromRooms(t *testing.T) {
	t.Run("删除客户端并清理空房间", func(t *testing.T) {
		repo := &mockRoomRepo{}
		cache := &mockRoomCache{}
		rm := newTestRoomManager(repo, cache)
		client := newTestClient("u1", "r1")

		rm.mu.Lock()
		rm.rooms["r1"] = map[*Client]bool{client: true}
		rm.mu.Unlock()

		rm.RemoveFromRooms(context.Background(), client)

		rm.mu.RLock()
		if _, ok := rm.rooms["r1"]; ok {
			t.Error("empty room r1 should be deleted from rm.rooms")
		}
		rm.mu.RUnlock()

		cache.mu.Lock()
		if len(cache.leftRooms) != 1 || cache.leftRooms[0].roomID != "r1" {
			t.Errorf("LeaveRoom calls = %v, want [r1]", cache.leftRooms)
		}
		cache.mu.Unlock()
	})

	t.Run("保留非空房间", func(t *testing.T) {
		repo := &mockRoomRepo{}
		cache := &mockRoomCache{}
		rm := newTestRoomManager(repo, cache)
		client1 := newTestClient("u1", "r1")
		client2 := newTestClient("u2", "r1")

		rm.mu.Lock()
		rm.rooms["r1"] = map[*Client]bool{client1: true, client2: true}
		rm.mu.Unlock()

		rm.RemoveFromRooms(context.Background(), client1)

		rm.mu.RLock()
		clients, ok := rm.rooms["r1"]
		if !ok {
			t.Fatal("room r1 should still exist with remaining client")
		}
		if !clients[client2] {
			t.Error("client2 should still be in room r1")
		}
		if clients[client1] {
			t.Error("client1 should have been removed from room r1")
		}
		rm.mu.RUnlock()
	})
}

func TestBroadcastToRoom(t *testing.T) {
	t.Run("发送到房间客户端", func(t *testing.T) {
		repo := &mockRoomRepo{}
		cache := &mockRoomCache{}
		rm := newTestRoomManager(repo, cache)
		client := newTestClient("u1", "r1")

		rm.mu.Lock()
		rm.rooms["r1"] = map[*Client]bool{client: true}
		rm.mu.Unlock()

		rm.BroadcastToRoom("r1", "hello")

		select {
		case msg := <-client.Send:
			if msg != "hello" {
				t.Errorf("received %v, want 'hello'", msg)
			}
		default:
			t.Error("expected message in client.Send, got nothing")
		}
	})

	t.Run("channel满时不阻塞", func(t *testing.T) {
		repo := &mockRoomRepo{}
		cache := &mockRoomCache{}
		rm := newTestRoomManager(repo, cache)
		client := newTestClientWithUnbufferedSend("u1")

		rm.mu.Lock()
		rm.rooms["r1"] = map[*Client]bool{client: true}
		rm.mu.Unlock()

		done := make(chan struct{})
		go func() {
			rm.BroadcastToRoom("r1", "msg")
			close(done)
		}()

		select {
		case <-done:
		case <-make(chan struct{}, 1):
			t.Fatal("BroadcastToRoom blocked on full channel")
		}
	})

	t.Run("不存在的房间不panic", func(t *testing.T) {
		repo := &mockRoomRepo{}
		cache := &mockRoomCache{}
		rm := newTestRoomManager(repo, cache)

		rm.BroadcastToRoom("nonexistent", "msg")
	})
}

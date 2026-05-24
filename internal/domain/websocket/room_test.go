package websocket

import (
	"context"
	"errors"
	"testing"
)

func TestLoadRooms_WritesToClientAndManager(t *testing.T) {
	repo := &mockRoomRepo{rooms: []string{"r1", "r2"}}
	cache := &mockRoomCache{}
	rm := newTestRoomManager(repo, cache)
	client := newTestClient("u1")

	rm.LoadRooms(context.Background(), client)

	if !client.Rooms["r1"] || !client.Rooms["r2"] {
		t.Errorf("client.Rooms = %v, want r1 and r2", client.Rooms)
	}
	rm.mu.RLock()
	if _, ok := rm.rooms["r1"][client]; !ok {
		t.Error("rm.rooms[r1][client] should be true")
	}
	if _, ok := rm.rooms["r2"][client]; !ok {
		t.Error("rm.rooms[r2][client] should be true")
	}
	rm.mu.RUnlock()

	cache.mu.Lock()
	if len(cache.joinedRooms) != 2 {
		t.Errorf("JoinRoom called %d times, want 2", len(cache.joinedRooms))
	}
	cache.mu.Unlock()
}

func TestLoadRooms_RepoErrorSkipsLoading(t *testing.T) {
	repo := &mockRoomRepo{roomsErr: errors.New("db down")}
	cache := &mockRoomCache{}
	rm := newTestRoomManager(repo, cache)
	client := newTestClient("u1")

	rm.LoadRooms(context.Background(), client)

	if len(client.Rooms) != 0 {
		t.Errorf("client.Rooms should be empty on error, got %v", client.Rooms)
	}
	cache.mu.Lock()
	if len(cache.joinedRooms) != 0 {
		t.Errorf("JoinRoom should not be called on error, called %d times", len(cache.joinedRooms))
	}
	cache.mu.Unlock()
}

func TestRemoveFromRooms_DeletesClientAndCleansEmptyRooms(t *testing.T) {
	repo := &mockRoomRepo{}
	cache := &mockRoomCache{}
	rm := newTestRoomManager(repo, cache)
	client := newTestClient("u1", "r1")

	// 先手动加入
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
}

func TestRemoveFromRooms_KeepsNonEmptyRooms(t *testing.T) {
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
}

func TestBroadcastToRoom_SendsToClients(t *testing.T) {
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
}

func TestBroadcastToRoom_NonBlockingWhenChannelFull(t *testing.T) {
	repo := &mockRoomRepo{}
	cache := &mockRoomCache{}
	rm := newTestRoomManager(repo, cache)
	client := newTestClientWithUnbufferedSend("u1")

	rm.mu.Lock()
	rm.rooms["r1"] = map[*Client]bool{client: true}
	rm.mu.Unlock()

	// 无缓冲 channel，写入会阻塞，但 BroadcastToRoom 用 select default 应该快速返回
	done := make(chan struct{})
	go func() {
		rm.BroadcastToRoom("r1", "msg")
		close(done)
	}()

	select {
	case <-done:
		// 成功，未阻塞
	case <-make(chan struct{}, 1):
		t.Fatal("BroadcastToRoom blocked on full channel")
	}
}

func TestBroadcastToRoom_NonExistentRoom(t *testing.T) {
	repo := &mockRoomRepo{}
	cache := &mockRoomCache{}
	rm := newTestRoomManager(repo, cache)

	// 不应该 panic
	rm.BroadcastToRoom("nonexistent", "msg")
}

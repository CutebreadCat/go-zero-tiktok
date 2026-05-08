package websocket

import (
	"context"
	"sync"

	"go_zero-tiktok/internal/types"

	"github.com/gorilla/websocket"
)

// CacheService 是缓存层的依赖抽象
type CacheService interface {
	SetOnline(ctx context.Context, userID string, addr string) error
	SetOffline(ctx context.Context, userID string, addr string) error
	AddMessage(ctx context.Context, message *types.MessageChat) (string, error)
	IncrUnread(ctx context.Context, userID, roomID string) error
}

// ChatRepository 是数据访问层的依赖抽象
type ChatRepository interface {
	GetJoinRooms(ctx context.Context, userID string) ([]string, error)
	StoreChatMessage(ctx context.Context, message *types.MessageChat) error
	GetChatRoomUsers(ctx context.Context, roomID string) ([]string, error)
}

type Hub struct {
	Clients map[*Client]bool
	Mu      sync.RWMutex
	Rooms   map[string]map[*Client]bool
	Cache   CacheService
	Chat    ChatRepository
}

type Client struct {
	Hub    *Hub
	UserID string
	Conn   *websocket.Conn
	Send   chan any
	Rooms  map[string]bool
	Cmu    sync.Mutex
}

func NewHub(cache CacheService, chat ChatRepository) *Hub {
	return &Hub{
		Clients: make(map[*Client]bool),
		Rooms:   make(map[string]map[*Client]bool),
		Cache:   cache,
		Chat:    chat,
	}
}

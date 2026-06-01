package websocket

import (
	"sync"

	"go_zero-tiktok/internal/types"

	"github.com/gorilla/websocket"
)

// ==================== 模型定义 ====================

// Limiter 限流器接口
type Limiter interface {
	Allow(userId string) (bool, error)
}

// Message WebSocket 消息结构
type Message struct {
	Message types.MessageChat `json:"message"`
	Typek   string            `json:"typek"`
}

// Hub 核心管理器
type Hub struct {
	presence PresenceManager
	rooms    RoomManager
	messages MessageManager
	limiter  Limiter
}

// Client WebSocket 客户端
type Client struct {
	Hub    *Hub
	UserID string
	Conn   *websocket.Conn
	Send   chan any
	Rooms  map[string]bool
	Cmu    sync.Mutex
}

// ==================== Hub 方法 ====================

func (h *Hub) Presence() PresenceManager { return h.presence }
func (h *Hub) Rooms() RoomManager        { return h.rooms }
func (h *Hub) Messages() MessageManager  { return h.messages }

// SetWriter 注入消息写入器（MQ Producer）
func (h *Hub) SetWriter(writer MessageWriter) {
	if mm, ok := h.messages.(*messageManager); ok {
		mm.writer = writer
	}
}

// ==================== 构造函数 ====================

func NewHub(pc PresenceCache, rc RoomCache, mc MessageCache, rr RoomRepository, mr MessageRepository, ai *AIChat, limiter Limiter) *Hub {
	rm := &roomManager{
		rooms: make(map[string]map[*Client]bool),
		repo:  rr,
		cache: rc,
	}
	pm := &presenceManager{
		clients: make(map[*Client]bool),
		cache:   pc,
		rooms:   rm,
	}
	mm := &messageManager{
		cache:    mc,
		repo:     mr,
		roomRepo: rr,
		rooms:    rm,
		ai:       ai,
	}
	return &Hub{
		presence: pm,
		rooms:    rm,
		messages: mm,
		limiter:  limiter,
	}
}

func NewClient(hub *Hub, userID string, conn *websocket.Conn) *Client {
	return &Client{
		Hub:    hub,
		UserID: userID,
		Conn:   conn,
		Send:   make(chan any, 256),
		Rooms:  make(map[string]bool),
	}
}

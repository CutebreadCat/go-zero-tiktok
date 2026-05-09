package websocket

import (
	"sync"

	"go_zero-tiktok/internal/types"

	"github.com/gorilla/websocket"
)

// ==================== 核心结构体 ====================

type Message struct {
	Message types.MessageChat `json:"message"`
	Typek   string            `json:"typek"`
}

type Hub struct {
	presence PresenceManager
	rooms    RoomManager
	messages MessageManager
}

type Client struct {
	Hub    *Hub
	UserID string
	Conn   *websocket.Conn
	Send   chan any
	Rooms  map[string]bool
	Cmu    sync.Mutex
}

func (h *Hub) Presence() PresenceManager { return h.presence }
func (h *Hub) Rooms() RoomManager        { return h.rooms }
func (h *Hub) Messages() MessageManager  { return h.messages }

func NewHub(pc PresenceCache, rc RoomCache, mc MessageCache, rr RoomRepository, mr MessageRepository) *Hub {
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
	}
	return &Hub{
		presence: pm,
		rooms:    rm,
		messages: mm,
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

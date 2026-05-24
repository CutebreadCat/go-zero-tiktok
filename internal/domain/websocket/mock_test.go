package websocket

import (
	"context"
	"sync"

	mqcontract "go_zero-tiktok/internal/shared/mq"
	"go_zero-tiktok/internal/types"

	"github.com/sashabaranov/go-openai"
)

// ==================== Mock RoomRepository ====================

type mockRoomRepo struct {
	rooms    []string
	users    []string
	roomsErr error
	usersErr error
}

func (m *mockRoomRepo) GetJoinRooms(_ context.Context, _ string) ([]string, error) {
	return m.rooms, m.roomsErr
}

func (m *mockRoomRepo) GetChatRoomUsers(_ context.Context, _ string) ([]string, error) {
	return m.users, m.usersErr
}

// ==================== Mock RoomCache ====================

type mockRoomCache struct {
	mu          sync.Mutex
	joinedRooms []struct{ roomID, userID string }
	leftRooms   []struct{ roomID, userID string }
	joinErr     error
	leaveErr    error
}

func (m *mockRoomCache) JoinRoom(_ context.Context, roomID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.joinedRooms = append(m.joinedRooms, struct{ roomID, userID string }{roomID, userID})
	return m.joinErr
}

func (m *mockRoomCache) LeaveRoom(_ context.Context, roomID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.leftRooms = append(m.leftRooms, struct{ roomID, userID string }{roomID, userID})
	return m.leaveErr
}

func (m *mockRoomCache) RoomHeartBeat(_ context.Context, _ string) error {
	return nil
}

// ==================== Mock MessageCache ====================

type mockMessageCache struct {
	addMsgID  string
	addMsgErr error

 incrErr error

	messages      []CacheMessage
	messagesErr   error

	unreadCount   int64
	unreadCountErr error

	clearErr error

	incrAIErr      error
	aiCount        int64
	aiCountErr     error
	clearAIErr     error
	addAIMsgID     string
	addAIMsgErr    error
	aiMessages     []CacheMessage
	aiMessagesErr  error
	clearAIStreamErr error

 mu              sync.Mutex
 incrUnreadCalls []struct{ userID, roomID string }
}

func (m *mockMessageCache) AddMessage(_ context.Context, _ *types.MessageChat) (string, error) {
	return m.addMsgID, m.addMsgErr
}

func (m *mockMessageCache) IncrUnread(_ context.Context, userID, roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.incrUnreadCalls = append(m.incrUnreadCalls, struct{ userID, roomID string }{userID, roomID})
	return m.incrErr
}

func (m *mockMessageCache) GetMessages(_ context.Context, _, _ string, _ int64) ([]CacheMessage, error) {
	return m.messages, m.messagesErr
}

func (m *mockMessageCache) GetUnreadCount(_ context.Context, _, _ string) (int64, error) {
	return m.unreadCount, m.unreadCountErr
}

func (m *mockMessageCache) ClearUnread(_ context.Context, _, _ string) error {
	return m.clearErr
}

func (m *mockMessageCache) IncrAIMessage(_ context.Context, _, _ string) error {
	return m.incrAIErr
}

func (m *mockMessageCache) GetAIMessageCount(_ context.Context, _, _ string) (int64, error) {
	return m.aiCount, m.aiCountErr
}

func (m *mockMessageCache) ClearAIMessage(_ context.Context, _, _ string) error {
	return m.clearAIErr
}

func (m *mockMessageCache) AddAIMessage(_ context.Context, _, _ string, _ *types.MessageChat) (string, error) {
	return m.addAIMsgID, m.addAIMsgErr
}

func (m *mockMessageCache) GetAIMessages(_ context.Context, _, _ string, _ int64) ([]CacheMessage, error) {
	return m.aiMessages, m.aiMessagesErr
}

func (m *mockMessageCache) ClearAIStream(_ context.Context, _, _ string) error {
	return m.clearAIStreamErr
}

// ==================== Mock MessageRepository ====================

type mockMessageRepo struct {
	storeErr error
}

func (m *mockMessageRepo) StoreChatMessage(_ context.Context, _ *types.MessageChat) error {
	return m.storeErr
}

// ==================== Mock MessageWriter ====================

type mockMessageWriter struct {
	mu       sync.Mutex
	events   []*mqcontract.Event
	sendErr  error
}

func (m *mockMessageWriter) SendMessage(_ context.Context, event *mqcontract.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return m.sendErr
}

// chanWriter 用 channel 收集事件，方便测试异步场景
type chanWriter struct {
	ch      chan *mqcontract.Event
	sendErr error
}

func (w *chanWriter) SendMessage(_ context.Context, event *mqcontract.Event) error {
	select {
	case w.ch <- event:
	default:
	}
	return w.sendErr
}

// ==================== Mock AIChatMessage (Agent) ====================

type mockAIAgent struct {
	reply    string
	runErr   error
	parsed   []openai.ChatCompletionMessage
}

func (m *mockAIAgent) Run(_ context.Context, _ string, _ []openai.ChatCompletionMessage) (string, error) {
	return m.reply, m.runErr
}

func (m *mockAIAgent) ParseMessageToOpenAIList(_ context.Context, msg []CacheMessage) []openai.ChatCompletionMessage {
	if m.parsed != nil {
		return m.parsed
	}
	return []openai.ChatCompletionMessage{}
}

// ==================== Helper: 创建测试用 Client ====================

func newTestClient(userID string, roomIDs ...string) *Client {
	rooms := make(map[string]bool)
	for _, r := range roomIDs {
		rooms[r] = true
	}
	return &Client{
		UserID: userID,
		Send:   make(chan any, 256),
		Rooms:  rooms,
	}
}

// newTestClientWithUnbufferedSend 创建 Send channel 无缓冲的 Client
func newTestClientWithUnbufferedSend(userID string) *Client {
	return &Client{
		UserID: userID,
		Send:   make(chan any),
		Rooms:  map[string]bool{"room1": true},
	}
}

// ==================== Helper: 创建 roomManager ====================

func newTestRoomManager(repo RoomRepository, cache RoomCache) *roomManager {
	return &roomManager{
		rooms: make(map[string]map[*Client]bool),
		repo:  repo,
		cache: cache,
	}
}

// ==================== Helper: 创建 messageManager ====================

func newTestMessageManager(
	cache MessageCache,
	repo MessageRepository,
	roomRepo RoomRepository,
	rooms RoomManager,
	writer MessageWriter,
	ai *AIChat,
) *messageManager {
	return &messageManager{
		cache:    cache,
		repo:     repo,
		roomRepo: roomRepo,
		rooms:    rooms,
		writer:   writer,
		ai:       ai,
	}
}

package websocket

import "context"

// ==================== 接口定义 ====================

// PresenceCache 在线状态缓存接口
type PresenceCache interface {
	SetOnline(ctx context.Context, userID string, addr string) error
	SetOffline(ctx context.Context, userID string, addr string) error
	HeartBeat(ctx context.Context, userID string, addr string) error
}

// PresenceManager 客户端连接管理接口
type PresenceManager interface {
	AddClient(ctx context.Context, client *Client)
	RemoveClient(ctx context.Context, client *Client)
}

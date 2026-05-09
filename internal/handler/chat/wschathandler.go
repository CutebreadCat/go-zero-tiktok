// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"net/http"

	"go_zero-tiktok/internal/svc"
	myutils "go_zero-tiktok/internal/utils"

	mywebsocket "go_zero-tiktok/internal/domain/websocket"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

func WsChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 解析请求参数 (如果需要)

		// 如果有表单或查询参数，可以在这里解析
		// if err := httpx.Parse(r, &req); err != nil { ... }

		// 2. 获取 Context
		ctx := r.Context()

		// 3. 直接在这里进行 WebSocket 升级
		// 这里使用了你之前定义的 upgrader 逻辑
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 生产环境请严格校验 Origin
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// 注意：Upgrade 失败后，不要尝试写 JSON，因为可能已经写了一半的 HTTP Header
			// 简单的记录日志即可
			logx.Errorf("WebSocket upgrade failed: %v", err)
			return
		}

		// 4. 升级成功，获取 UserID (根据你的鉴权方式调整)
		// 假设你是从 Token 或者 Query 中获取的
		userid, err := myutils.GetUserIDFromContext(ctx) // 你需要实现这个函数来从请求中提取用户ID
		if err != nil {
			logx.Errorf("Failed to get user ID from context: %v", err)
			conn.Close() // 获取用户ID失败，关闭 WebSocket 连接
			return
		}
		if userid == "" {
			conn.Close() // 如果没有用户ID，关闭 WebSocket 连接
			return
		}

		// 5. 创建 Client 并注册到 Hub
		client := mywebsocket.NewClient(svcCtx.Hub, userid, conn)
		svcCtx.Hub.Presence().AddClient(ctx, client)

		// 6. 启动读写协程
		go client.ReadLoop()
		go client.WriteLoop()

		// 7. Handler 结束，连接已由 Client 的 goroutine 接管
	}
}

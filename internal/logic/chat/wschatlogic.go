// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"net/http"
	"sync"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	mywebsocket "go_zero-tiktok/internal/websocket"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type WsChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWsChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WsChatLogic {
	return &WsChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *WsChatLogic) WsChat(req *types.WsChatRequest) (err error) {
	// todo: add your logic here and delete this line
	var userid string
	userid, err = myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败: %v", err)
		return err
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(l.ctx.Value("httpResponseWriter").(http.ResponseWriter), l.ctx.Value("httpRequest").(*http.Request), nil)
	if err != nil {
		l.Logger.Errorf("升级WebSocket连接失败: %v", err)
		return err
	}
	client := &mywebsocket.Client{
		Hub:    l.svcCtx.Hub,
		UserID: userid,
		Send:   make(chan any, 256),
		Rooms:  make(map[string]bool),
		Conn:   conn,
		Cmu:    sync.Mutex{},
	}
	l.svcCtx.Hub.AddClient(l.ctx, client)
	go client.ReadLoop()
	go client.WriteLoop()
	return
}

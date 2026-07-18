package websocket

import (
	"context"
	appLogger "go_zero-tiktok/Prometheus/logger"
	"time"

	myutils "go_zero-tiktok/pkg/utils"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

func (c *Client) ReadLoop() {
	ctx, hand := context.WithTimeout(context.Background(), clientContextTTL)
	defer func() {
		c.Hub.Presence().RemoveClient(ctx, c)
		c.Conn.Close()
		hand()
	}()
	c.Conn.SetReadLimit(clientReadLimit)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { _ = c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var msg Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			break
		}

		if allowed, err := c.Hub.limiter.Allow(c.UserID); err != nil {
			appLogger.Infof("闄愭祦妫€鏌ュけ璐? %v", err)
			continue
		} else if !allowed {
			c.Send <- map[string]string{typeKey: MessageTypeError, rateLimitMessageKey: wsRateLimitMessage}
			continue
		}

		msg.Message.ID = uuid.New().String()
		msg.Message.SenderID = c.UserID
		msg.Message.CreatedAt = myutils.TsToStr(time.Now().Unix(), dateTimeLayout)
		switch msg.Typek {
		case EventTypeMessage:
			c.Hub.Messages().HandleMessage(ctx, c, &msg.Message)
		case EventTypeUnread:
			c.Hub.Messages().HandleGetUnread(ctx, c, msg.Message.RoomID)
		case MessageTypePing:
			c.Send <- map[string]string{"req": MessageTypePong}
		default:
			appLogger.Infof("鏈煡鐨勬秷鎭被鍨? %s", msg.Typek)
			continue
		}
	}
}

func (c *Client) WriteLoop() {
	ticker := time.NewTicker(pongWait / 2)
	defer ticker.Stop()
	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(pongWait))
			c.Cmu.Lock()
			if !ok {
				appLogger.Info("鍙戦€侀€氶亾宸插叧闂紝閫€鍑?WriteLoop")
				c.Cmu.Unlock()
				return
			}
			if err := c.Conn.WriteJSON(message); err != nil {
				appLogger.Infof("鍐欏叆娑堟伅鍒板鎴风 %s 澶辫触: %v", c.UserID, err)
				c.Cmu.Unlock()
				return
			}
			c.Cmu.Unlock()
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(pongWait))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				appLogger.Infof("鍙戦€?ping 鍒板鎴风 %s 澶辫触: %v", c.UserID, err)
				return
			}
		}
	}
}

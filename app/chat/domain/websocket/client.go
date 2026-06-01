package websocket

import (
	"context"
	"log"
	"time"

	myutils "go_zero-tiktok/internal/utils"

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
			log.Printf("限流检查失败: %v", err)
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
			log.Printf("未知的消息类型: %s", msg.Typek)
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
				log.Println("发送通道已关闭，退出 WriteLoop")
				c.Cmu.Unlock()
				return
			}
			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("写入消息到客户端 %s 失败: %v", c.UserID, err)
				c.Cmu.Unlock()
				return
			}
			c.Cmu.Unlock()
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(pongWait))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				log.Printf("发送 ping 到客户端 %s 失败: %v", c.UserID, err)
				return
			}
		}
	}
}

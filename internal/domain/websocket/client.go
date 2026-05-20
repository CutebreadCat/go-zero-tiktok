package websocket

import (
	"context"
	"fmt"
	"time"

	myutils "go_zero-tiktok/internal/utils"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

func (c *Client) ReadLoop() {
	ctx, hand := context.WithTimeout(context.Background(), time.Hour*24)
	defer func() {
		c.Hub.Presence().RemoveClient(ctx, c)
		c.Conn.Close()
		hand()
	}()
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var msg Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			break
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
			fmt.Printf("未知的消息类型: %s\n", msg.Typek)
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
			c.Conn.SetWriteDeadline(time.Now().Add(pongWait))
			c.Cmu.Lock()
			if !ok {
				fmt.Printf("发送通道已关闭，退出 WriteLoop\n")
				c.Cmu.Unlock()
				return
			}
			if err := c.Conn.WriteJSON(message); err != nil {
				fmt.Printf("写入消息到客户端 %s 失败: %v\n", c.UserID, err)
				c.Cmu.Unlock()
				return
			}
			c.Cmu.Unlock()
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(pongWait))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				fmt.Printf("发送 ping 到客户端 %s 失败: %v\n", c.UserID, err)
				return
			}
		}
	}
}

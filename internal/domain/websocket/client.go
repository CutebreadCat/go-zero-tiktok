package websocket

import (
	"context"
	"log"
	"time"

	myutils "go_zero-tiktok/internal/utils"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

var pongWait = 60 * time.Second

// ==================== Client 读写循环 ====================

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
		msg.Message.CreatedAt = myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05")
		switch msg.Typek {
		case "message":
			c.Hub.Messages().HandleMessage(ctx, c, &msg.Message)
		case "get_unread":
			c.Hub.Messages().HandleGetUnread(ctx, c, msg.Message.RoomID)
		case "ping":
			c.Send <- map[string]string{"req": "pong"}
		default:
			log.Printf("Unknown message type: %s", msg.Typek)
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
				log.Println("Send channel closed, exiting WriteLoop")
				c.Cmu.Unlock()
				return
			}
			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("Failed to write message to client %s: %v", c.UserID, err)
				c.Cmu.Unlock()
				return
			}
			c.Cmu.Unlock()
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(pongWait))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping to client %s: %v", c.UserID, err)
				return
			}
		}
	}
}

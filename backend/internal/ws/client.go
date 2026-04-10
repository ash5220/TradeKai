package ws

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client represents a single WebSocket connection.
type Client struct {
	id     uuid.UUID
	userID uuid.UUID
	hub    *Hub
	conn   *websocket.Conn
	send   chan Message
}

// ID returns the unique connection identifier.
func (c *Client) ID() uuid.UUID { return c.id }

// UserID returns the authenticated user associated with this connection.
func (c *Client) UserID() uuid.UUID { return c.userID }

// readPump listens for incoming client messages (subscribe/unsubscribe commands).
// It runs in its own goroutine and closes the connection on error.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.log.Warn("ws unexpected close", zap.Error(err),
					zap.Stringer("client", c.id))
			}
			return
		}
		c.handleClientMessage(raw)
	}
}

// writePump drains the send channel and writes messages to the WebSocket.
// It also sends periodic pings to keep the connection alive.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// clientMsg is the structure of messages sent by the client.
type clientMsg struct {
	Action string `json:"action"` // "subscribe" | "unsubscribe"
	Room   string `json:"room"`
}

func (c *Client) handleClientMessage(raw []byte) {
	var m clientMsg
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	switch m.Action {
	case "subscribe":
		if !c.canAccessRoom(m.Room) {
			return
		}
		c.hub.JoinRoom(c, m.Room)
	case "unsubscribe":
		if !c.canAccessRoom(m.Room) {
			return
		}
		c.hub.LeaveRoom(c, m.Room)
	}
}

func (c *Client) canAccessRoom(room string) bool {
	if strings.HasPrefix(room, "ticks:") {
		return strings.TrimSpace(strings.TrimPrefix(room, "ticks:")) != ""
	}
	return room == fmt.Sprintf("orders:%s", c.userID)
}

// Package ws manages WebSocket connections with room-based pub/sub.
package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   1024,
	WriteBufferSize:  4096,
	CheckOrigin: func(r *http.Request) bool {
		// Origin validation is handled by CORS middleware upstream
		return true
	},
}

// Message is a JSON-serialisable message sent to WebSocket clients.
type Message struct {
	Type    string `json:"type"`
	Room    string `json:"room,omitempty"`
	Payload any    `json:"payload"`
}

// Hub manages all WebSocket client connections and room subscriptions.
type Hub struct {
	log *zap.Logger

	mu      sync.RWMutex
	clients map[uuid.UUID]*Client      // connID → client
	rooms   map[string]map[uuid.UUID]*Client // room → connID → client

	register   chan *Client
	unregister chan *Client
	broadcast  chan roomMessage
}

type roomMessage struct {
	room string
	msg  Message
}

// NewHub creates and returns a Hub ready to run.
func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		log:        log,
		clients:    make(map[uuid.UUID]*Client),
		rooms:      make(map[string]map[uuid.UUID]*Client),
		register:   make(chan *Client, 32),
		unregister: make(chan *Client, 32),
		broadcast:  make(chan roomMessage, 256),
	}
}

// Run processes registration, unregistration, and broadcast events.
// It must be called in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c.id] = c
			h.mu.Unlock()
			h.log.Debug("ws client registered", zap.Stringer("id", c.id))

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c.id]; ok {
				delete(h.clients, c.id)
				for room, members := range h.rooms {
					delete(members, c.id)
					if len(members) == 0 {
						delete(h.rooms, room)
					}
				}
				close(c.send)
			}
			h.mu.Unlock()
			h.log.Debug("ws client unregistered", zap.Stringer("id", c.id))

		case rm := <-h.broadcast:
			h.mu.RLock()
			members := h.rooms[rm.room]
			targets := make([]*Client, 0, len(members))
			for _, c := range members {
				targets = append(targets, c)
			}
			h.mu.RUnlock()

			for _, c := range targets {
				select {
				case c.send <- rm.msg:
				default:
					h.log.Warn("ws client send buffer full, dropping message",
						zap.Stringer("id", c.id), zap.String("room", rm.room))
				}
			}
		}
	}
}

// Publish sends a message to all clients subscribed to room.
func (h *Hub) Publish(room string, msg Message) {
	h.broadcast <- roomMessage{room: room, msg: msg}
}

// JoinRoom adds client c to the named room.
func (h *Hub) JoinRoom(c *Client, room string) {
	h.mu.Lock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[uuid.UUID]*Client)
	}
	h.rooms[room][c.id] = c
	h.mu.Unlock()
}

// LeaveRoom removes client c from the named room.
func (h *Hub) LeaveRoom(c *Client, room string) {
	h.mu.Lock()
	if members, ok := h.rooms[room]; ok {
		delete(members, c.id)
		if len(members) == 0 {
			delete(h.rooms, room)
		}
	}
	h.mu.Unlock()
}

// Upgrade upgrades an HTTP connection to WebSocket and registers the new client.
func (h *Hub) Upgrade(w http.ResponseWriter, r *http.Request, userID uuid.UUID) (*Client, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	c := &Client{
		id:     uuid.New(),
		userID: userID,
		hub:    h,
		conn:   conn,
		send:   make(chan Message, 256),
	}

	h.register <- c
	go c.writePump()
	go c.readPump()

	return c, nil
}

// ActiveConnections returns the number of currently registered clients.
func (h *Hub) ActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

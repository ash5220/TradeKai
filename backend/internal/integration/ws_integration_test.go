//go:build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/rashevskyv/tradekai/internal/auth"
	"github.com/rashevskyv/tradekai/internal/ws"
)

func TestIntegrationWebSocketSubscribeAndPublish(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	log := zap.NewNop()
	hub := ws.NewHub(log)
	go hub.Run()

	jwtManager := auth.NewManager("integration-secret-012345678901234567890", 15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()
	pair, err := jwtManager.Generate(userID, "ws@example.com")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	r := gin.New()
	wsHandler := wsHandlerForTest(hub, jwtManager)
	r.GET("/ws", wsHandler)

	srv := httptest.NewServer(r)
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/ws?token=" + pair.AccessToken
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer conn.Close()

	room := fmt.Sprintf("orders:%s", userID)
	subMsg := map[string]string{"action": "subscribe", "room": room}
	if err := conn.WriteJSON(subMsg); err != nil {
		t.Fatalf("WriteJSON(subscribe) error = %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	hub.Publish(room, ws.Message{Type: "order_update", Room: room, Payload: map[string]string{"status": "submitted"}})

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	var msg map[string]any
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got, ok := msg["type"].(string); !ok || got != "order_update" {
		t.Errorf("message type = %v, want order_update", msg["type"])
	}
}

func wsHandlerForTest(hub *ws.Hub, jwtManager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		claims, err := jwtManager.Validate(tokenStr)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		if _, err := hub.Upgrade(c.Writer, c.Request, claims.UserID); err != nil {
			c.AbortWithStatus(400)
			return
		}
	}
}

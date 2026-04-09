package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rashevskyv/tradekai/internal/auth"
	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/ws"
)

// WSHandler upgrades HTTP connections to WebSocket.
type WSHandler struct {
	hub        *ws.Hub
	jwtManager *auth.Manager
}

// NewWSHandler creates a WSHandler.
func NewWSHandler(hub *ws.Hub, jwtManager *auth.Manager) *WSHandler {
	return &WSHandler{hub: hub, jwtManager: jwtManager}
}

// Upgrade godoc
// @Summary Upgrade to WebSocket connection
// @Tags websocket
// @Param token query string true "JWT access token"
// @Router /ws [get]
func (h *WSHandler) Upgrade(c *gin.Context) {
	// JWT is passed via query param (can't set Authorization header in WebSocket)
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return
	}

	claims, err := h.jwtManager.Validate(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return
	}

	if _, err := h.hub.Upgrade(c.Writer, c.Request, claims.UserID); err != nil {
		// Upgrade writes its own error response on failure
		return
	}
}

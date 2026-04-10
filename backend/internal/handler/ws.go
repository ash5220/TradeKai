package handler

import (
	"net/http"
	"strings"

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
// @Param Authorization header string false "Bearer access token"
// @Param Sec-WebSocket-Protocol header string false "Include access-token.<JWT> as one subprotocol entry"
// @Router /ws [get]
func (h *WSHandler) Upgrade(c *gin.Context) {
	tokenStr := tokenFromRequest(c)
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

func tokenFromRequest(c *gin.Context) string {
	if token := tokenFromAuthorizationHeader(c.GetHeader("Authorization")); token != "" {
		return token
	}
	return tokenFromWSSubprotocolHeader(c.GetHeader("Sec-WebSocket-Protocol"))
}

func tokenFromAuthorizationHeader(v string) string {
	parts := strings.SplitN(strings.TrimSpace(v), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func tokenFromWSSubprotocolHeader(v string) string {
	for _, raw := range strings.Split(v, ",") {
		proto := strings.TrimSpace(raw)
		if !strings.HasPrefix(proto, "access-token.") {
			continue
		}
		token := strings.TrimPrefix(proto, "access-token.")
		if token != "" {
			return token
		}
	}
	return ""
}

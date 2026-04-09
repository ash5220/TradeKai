package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/rashevskyv/tradekai/internal/domain"
)

const (
	// ContextKeyUserID is the Gin context key for the authenticated user's UUID.
	ContextKeyUserID = "user_id"
	// ContextKeyEmail is the Gin context key for the authenticated user's email.
	ContextKeyEmail = "email"
)

// Middleware returns a Gin middleware that validates JWT tokens from the
// Authorization header and injects claims into the request context.
func Middleware(jwtManager *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "malformed authorization header"})
			return
		}

		claims, err := jwtManager.Validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Next()
	}
}

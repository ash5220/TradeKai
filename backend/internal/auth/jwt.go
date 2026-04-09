package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims are the JWT payload fields.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Manager handles JWT creation and validation.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager creates a JWT Manager.
func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Generate creates a new access + refresh token pair for the given user.
func (m *Manager) Generate(userID uuid.UUID, email string) (TokenPair, error) {
	access, err := m.sign(userID, email, m.accessTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign access token: %w", err)
	}

	refresh, err := m.sign(userID, email, m.refreshTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign refresh token: %w", err)
	}

	return TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

// Validate parses and validates a token string, returning its claims.
func (m *Manager) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (m *Manager) sign(userID uuid.UUID, email string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

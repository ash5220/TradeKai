package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered account.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

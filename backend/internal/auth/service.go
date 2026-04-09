package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/store/generated"
)

const bcryptCost = 12

// Service handles user authentication operations.
type Service struct {
	queries *generated.Queries
	jwt     *Manager
}

// NewService creates an auth Service.
func NewService(db *pgxpool.Pool, jwt *Manager) *Service {
	return &Service{
		queries: generated.New(db),
		jwt:     jwt,
	}
}

// Register creates a new user account and returns a token pair.
func (s *Service) Register(ctx context.Context, email, password string) (TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return TokenPair{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.queries.CreateUser(ctx, generated.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		// Postgres unique violation code 23505
		if isUniqueViolation(err) {
			return TokenPair{}, domain.ErrEmailAlreadyExists
		}
		return TokenPair{}, fmt.Errorf("create user: %w", err)
	}

	return s.jwt.Generate(user.ID, user.Email)
}

// Login verifies credentials and returns a token pair.
func (s *Service) Login(ctx context.Context, email, password string) (TokenPair, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return TokenPair{}, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return TokenPair{}, domain.ErrInvalidCredentials
	}

	return s.jwt.Generate(user.ID, user.Email)
}

// Refresh validates a refresh token and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.jwt.Validate(refreshToken)
	if err != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}

	// Confirm user still exists
	user, err := s.queries.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}

	return s.jwt.Generate(user.ID, user.Email)
}

// isUniqueViolation returns true when err is a PostgreSQL unique-constraint error.
func isUniqueViolation(err error) bool {
	return err != nil && containsCode(err.Error(), "23505")
}

func containsCode(msg, code string) bool {
	return len(msg) > 0 && (len(msg) >= len(code) &&
		(func() bool {
			for i := 0; i <= len(msg)-len(code); i++ {
				if msg[i:i+len(code)] == code {
					return true
				}
			}
			return false
		})())
}

package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rashevskyv/tradekai/internal/auth"
)

func TestJWT_RoundTrip(t *testing.T) {
	t.Helper()
	mgr := auth.NewManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	pair, err := mgr.Generate(userID, email)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("Generate() AccessToken is empty")
	}
	if pair.RefreshToken == "" {
		t.Error("Generate() RefreshToken is empty")
	}

	claims, err := mgr.Validate(pair.AccessToken)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("Validate() UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Validate() email = %v, want %v", claims.Email, email)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	t.Helper()
	mgr := auth.NewManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	_, err := mgr.Validate("not.a.valid.token")
	if err == nil {
		t.Error("Validate() with invalid token returned nil error, want error")
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	t.Helper()
	mgr := auth.NewManager("test-secret-key", -1*time.Second, 7*24*time.Hour) // expired immediately
	pair, err := mgr.Generate(uuid.New(), "x@x.com")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	_, err = mgr.Validate(pair.AccessToken)
	if err == nil {
		t.Error("Validate() with expired token returned nil error, want error")
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	t.Helper()
	mgr1 := auth.NewManager("secret-1", 15*time.Minute, time.Hour)
	mgr2 := auth.NewManager("secret-2", 15*time.Minute, time.Hour)

	pair, _ := mgr1.Generate(uuid.New(), "u@u.com")
	_, err := mgr2.Validate(pair.AccessToken)
	if err == nil {
		t.Error("Validate() with wrong secret returned nil error, want error")
	}
}

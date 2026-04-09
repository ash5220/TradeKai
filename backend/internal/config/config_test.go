package config

import (
	"strings"
	"testing"
	"time"
)

func baseConfig() Config {
	return Config{
		Server: ServerConfig{Mode: "development"},
		Database: DatabaseConfig{
			URL:            "postgres://tradekai:tradekai@localhost:5432/tradekai?sslmode=disable",
			MaxConnections: 25,
			MinConnections: 5,
		},
		JWT: JWTConfig{
			Secret:     strings.Repeat("a", 32),
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
		Log: LogConfig{Format: "console"},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:4200"},
		},
		Rate: RateLimitConfig{API: 100, Auth: 10},
		Market: MarketConfig{
			Provider: "simulated",
		},
		Order: OrderConfig{
			Executor: "simulated",
		},
	}
}

func TestValidate_ProductionRequiresJSONLogFormat(t *testing.T) {
	t.Helper()
	cfg := baseConfig()
	cfg.Server.Mode = "production"
	cfg.CORS.AllowedOrigins = []string{"https://app.example.com"}

	err := cfg.validate()
	if err == nil || !strings.Contains(err.Error(), "LOG_FORMAT") {
		t.Errorf("validate() error = %v, want LOG_FORMAT production error", err)
	}
}

func TestValidate_ProductionRejectsLocalhostCORS(t *testing.T) {
	t.Helper()
	cfg := baseConfig()
	cfg.Server.Mode = "production"
	cfg.Log.Format = "json"
	cfg.CORS.AllowedOrigins = []string{"http://localhost:4200"}

	err := cfg.validate()
	if err == nil || !strings.Contains(err.Error(), "localhost") {
		t.Errorf("validate() error = %v, want localhost CORS error", err)
	}
}

func TestValidate_RejectsShortJWTSecret(t *testing.T) {
	t.Helper()
	cfg := baseConfig()
	cfg.JWT.Secret = "short-secret"

	err := cfg.validate()
	if err == nil || !strings.Contains(err.Error(), "at least 32") {
		t.Errorf("validate() error = %v, want JWT length error", err)
	}
}

func TestValidate_RejectsWildcardCORS(t *testing.T) {
	t.Helper()
	cfg := baseConfig()
	cfg.CORS.AllowedOrigins = []string{"*"}

	err := cfg.validate()
	if err == nil || !strings.Contains(err.Error(), "wildcard") {
		t.Errorf("validate() error = %v, want wildcard CORS error", err)
	}
}

func TestValidate_AcceptsProductionHardenedConfig(t *testing.T) {
	t.Helper()
	cfg := baseConfig()
	cfg.Server.Mode = "production"
	cfg.Log.Format = "json"
	cfg.CORS.AllowedOrigins = []string{"https://app.example.com"}
	cfg.Rate = RateLimitConfig{API: 300, Auth: 30}

	if err := cfg.validate(); err != nil {
		t.Errorf("validate() error = %v, want nil", err)
	}
}

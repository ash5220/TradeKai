package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Alpaca   AlpacaConfig
	Risk     RiskConfig
	Log      LogConfig
	CORS     CORSConfig
	Rate     RateLimitConfig
	Market   MarketConfig
	Order    OrderConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"SERVER_PORT"`
	Mode string `mapstructure:"SERVER_MODE"`
}

type DatabaseConfig struct {
	URL            string `mapstructure:"DATABASE_URL"`
	MaxConnections int32  `mapstructure:"DATABASE_MAX_CONNECTIONS"`
	MinConnections int32  `mapstructure:"DATABASE_MIN_CONNECTIONS"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"JWT_SECRET"`
	AccessTTL  time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	RefreshTTL time.Duration `mapstructure:"JWT_REFRESH_TTL"`
}

type AlpacaConfig struct {
	APIKey    string `mapstructure:"ALPACA_API_KEY"`
	APISecret string `mapstructure:"ALPACA_API_SECRET"`
	BaseURL   string `mapstructure:"ALPACA_BASE_URL"`
	DataFeed  string `mapstructure:"ALPACA_DATA_FEED"`
}

type RiskConfig struct {
	MaxPositionSize        int           `mapstructure:"RISK_MAX_POSITION_SIZE"`
	MaxOpenOrders          int           `mapstructure:"RISK_MAX_OPEN_ORDERS"`
	DailyLossLimit         float64       `mapstructure:"RISK_DAILY_LOSS_LIMIT"`
	DuplicateTradeWindow   time.Duration `mapstructure:"RISK_DUPLICATE_TRADE_WINDOW"`
	MaxPortfolioExposure   float64       `mapstructure:"RISK_MAX_PORTFOLIO_EXPOSURE"`
}

type LogConfig struct {
	Level  string `mapstructure:"LOG_LEVEL"`
	Format string `mapstructure:"LOG_FORMAT"`
}

type CORSConfig struct {
	AllowedOrigins []string
}

type RateLimitConfig struct {
	API  int `mapstructure:"RATE_LIMIT_API"`
	Auth int `mapstructure:"RATE_LIMIT_AUTH"`
}

type MarketConfig struct {
	Provider string `mapstructure:"MARKET_DATA_PROVIDER"` // alpaca | simulated
}

type OrderConfig struct {
	Executor string `mapstructure:"ORDER_EXECUTOR"` // alpaca | simulated
}

// Load reads configuration from .env file and environment variables.
// Environment variables take precedence over file values.
func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SERVER_MODE", "development")
	v.SetDefault("DATABASE_MAX_CONNECTIONS", 25)
	v.SetDefault("DATABASE_MIN_CONNECTIONS", 5)
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "168h")
	v.SetDefault("ALPACA_BASE_URL", "https://paper-api.alpaca.markets")
	v.SetDefault("ALPACA_DATA_FEED", "iex")
	v.SetDefault("RISK_MAX_POSITION_SIZE", 100)
	v.SetDefault("RISK_MAX_OPEN_ORDERS", 10)
	v.SetDefault("RISK_DAILY_LOSS_LIMIT", 1000.0)
	v.SetDefault("RISK_DUPLICATE_TRADE_WINDOW", "60s")
	v.SetDefault("RISK_MAX_PORTFOLIO_EXPOSURE", 50000.0)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "console")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:4200")
	v.SetDefault("RATE_LIMIT_API", 100)
	v.SetDefault("RATE_LIMIT_AUTH", 10)
	v.SetDefault("MARKET_DATA_PROVIDER", "simulated")
	v.SetDefault("ORDER_EXECUTOR", "simulated")

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// .env file is optional; environment variables are sufficient
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{}

	cfg.Server.Port = v.GetInt("SERVER_PORT")
	cfg.Server.Mode = v.GetString("SERVER_MODE")

	cfg.Database.URL = v.GetString("DATABASE_URL")
	cfg.Database.MaxConnections = int32(v.GetInt("DATABASE_MAX_CONNECTIONS"))
	cfg.Database.MinConnections = int32(v.GetInt("DATABASE_MIN_CONNECTIONS"))

	cfg.JWT.Secret = v.GetString("JWT_SECRET")
	cfg.JWT.AccessTTL = v.GetDuration("JWT_ACCESS_TTL")
	cfg.JWT.RefreshTTL = v.GetDuration("JWT_REFRESH_TTL")

	cfg.Alpaca.APIKey = v.GetString("ALPACA_API_KEY")
	cfg.Alpaca.APISecret = v.GetString("ALPACA_API_SECRET")
	cfg.Alpaca.BaseURL = v.GetString("ALPACA_BASE_URL")
	cfg.Alpaca.DataFeed = v.GetString("ALPACA_DATA_FEED")

	cfg.Risk.MaxPositionSize = v.GetInt("RISK_MAX_POSITION_SIZE")
	cfg.Risk.MaxOpenOrders = v.GetInt("RISK_MAX_OPEN_ORDERS")
	cfg.Risk.DailyLossLimit = v.GetFloat64("RISK_DAILY_LOSS_LIMIT")
	cfg.Risk.DuplicateTradeWindow = v.GetDuration("RISK_DUPLICATE_TRADE_WINDOW")
	cfg.Risk.MaxPortfolioExposure = v.GetFloat64("RISK_MAX_PORTFOLIO_EXPOSURE")

	cfg.Log.Level = v.GetString("LOG_LEVEL")
	cfg.Log.Format = v.GetString("LOG_FORMAT")

	rawOrigins := v.GetString("CORS_ALLOWED_ORIGINS")
	cfg.CORS.AllowedOrigins = splitTrimmed(rawOrigins, ",")

	cfg.Rate.API = v.GetInt("RATE_LIMIT_API")
	cfg.Rate.Auth = v.GetInt("RATE_LIMIT_AUTH")

	cfg.Market.Provider = v.GetString("MARKET_DATA_PROVIDER")
	cfg.Order.Executor = v.GetString("ORDER_EXECUTOR")

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.JWT.AccessTTL <= 0 {
		return fmt.Errorf("JWT_ACCESS_TTL must be positive")
	}
	if c.JWT.RefreshTTL <= 0 {
		return fmt.Errorf("JWT_REFRESH_TTL must be positive")
	}
	validProviders := map[string]bool{"alpaca": true, "simulated": true}
	if !validProviders[c.Market.Provider] {
		return fmt.Errorf("MARKET_DATA_PROVIDER must be one of: alpaca, simulated")
	}
	validExecutors := map[string]bool{"alpaca": true, "simulated": true}
	if !validExecutors[c.Order.Executor] {
		return fmt.Errorf("ORDER_EXECUTOR must be one of: alpaca, simulated")
	}
	return nil
}

func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

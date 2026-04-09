package domain

import "errors"

// Sentinel domain errors. Use errors.Is / errors.As for checking.
var (
	ErrNotFound             = errors.New("not found")
	ErrDuplicateOrder       = errors.New("duplicate order")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrMaxPositionExceeded  = errors.New("max position size exceeded")
	ErrMaxOpenOrders        = errors.New("max open orders exceeded")
	ErrDailyLossExceeded    = errors.New("daily loss limit exceeded")
	ErrPortfolioExposure    = errors.New("max portfolio exposure exceeded")
	ErrOrderAlreadyTerminal = errors.New("order is already in a terminal state")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrSymbolNotFound       = errors.New("symbol not found")
	ErrStrategyNotFound     = errors.New("strategy not found")
	ErrStrategyAlreadyRunning = errors.New("strategy is already running")
)

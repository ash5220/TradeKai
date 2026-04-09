// Package domain defines all port interfaces used by the application.
// Interfaces are defined here (consumer side) and implemented by adapter packages.
package domain

import "context"

// MarketDataProvider is the port for receiving real-time market data.
// Implemented by: market.AlpacaProvider, market.SimulatedProvider
type MarketDataProvider interface {
	// Connect establishes the data stream for the given symbols.
	Connect(ctx context.Context, symbols []string) error
	// Subscribe returns a channel that receives ticks for a symbol.
	Subscribe(symbol string) (<-chan Tick, error)
	// Close shuts down the data stream and releases resources.
	Close() error
}

// OrderExecutor is the port for submitting orders to an exchange.
// Implemented by: order.AlpacaExecutor, order.SimulatedExecutor
type OrderExecutor interface {
	// PlaceOrder submits the order and returns the exchange-assigned ID.
	PlaceOrder(ctx context.Context, order Order) (string, error)
	// CancelOrder requests cancellation of the order at the exchange.
	CancelOrder(ctx context.Context, orderID string) error
	// GetOrderStatus queries the current status from the exchange.
	GetOrderStatus(ctx context.Context, orderID string) (OrderStatus, error)
}

// Strategy is the port for a trading algorithm.
// Implemented by: strategy.RSIStrategy, strategy.MACDCrossoverStrategy
type Strategy interface {
	// Name returns a unique, human-readable identifier for the strategy.
	Name() string
	// RequiredIndicators returns the set of indicator names this strategy needs.
	RequiredIndicators() []Indicator
	// Evaluate runs the strategy logic given a completed candle and the current
	// indicator values (keyed by indicator name). Returns a TradeSignal.
	Evaluate(candle Candle, indicators map[string]float64) TradeSignal
}

// Indicator is the port for a streaming technical indicator.
// Implemented by: strategy/indicator.SMA, EMA, RSI, MACD
type Indicator interface {
	// Name returns a unique identifier, e.g. "RSI(14)".
	Name() string
	// Add feeds the next candle into the indicator and returns the updated value.
	// Returns false if there is not yet enough data to produce a value.
	Add(candle Candle) (float64, bool)
	// Value returns the most recent computed value without advancing state.
	// Returns false if no value has been computed yet.
	Value() (float64, bool)
	// Reset clears all internal state.
	Reset()
}

// RiskRule is the port for a pre-trade risk check.
// Implemented by: risk.MaxPositionRule, risk.DailyLossRule, etc.
type RiskRule interface {
	// Name identifies the rule for logging.
	Name() string
	// Check evaluates the proposed order against the current portfolio.
	// Returns a domain error (e.g. ErrMaxPositionExceeded) if the order should
	// be rejected, or nil if it passes.
	Check(ctx context.Context, order Order, portfolio PortfolioSummary) error
}

// Package risk contains pre-trade risk rules.
package risk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// MaxPositionRule rejects orders that would breach a max-shares-per-symbol limit.
type MaxPositionRule struct {
	maxQty float64
}

func NewMaxPositionRule(maxQty float64) *MaxPositionRule { return &MaxPositionRule{maxQty} }
func (r *MaxPositionRule) Name() string                  { return "MaxPosition" }

func (r *MaxPositionRule) Check(_ context.Context, order domain.Order, portfolio domain.PortfolioSummary) error {
	for _, pos := range portfolio.Positions {
		if pos.Symbol == order.Symbol {
			projected := pos.Qty
			if order.Side == domain.OrderSideBuy {
				projected += order.Qty
			} else {
				projected -= order.Qty
			}
			if projected > r.maxQty {
				return fmt.Errorf("%w: %.2f > max %.2f for %s",
					domain.ErrMaxPositionExceeded, projected, r.maxQty, order.Symbol)
			}
			return nil
		}
	}
	// No existing position — check the qty directly
	if order.Side == domain.OrderSideBuy && order.Qty > r.maxQty {
		return fmt.Errorf("%w: %.2f > max %.2f for %s",
			domain.ErrMaxPositionExceeded, order.Qty, r.maxQty, order.Symbol)
	}
	return nil
}

// MaxOpenOrdersRule rejects new orders when a user has too many open orders.
type MaxOpenOrdersRule struct {
	max    int
	mu     sync.RWMutex
	counts map[string]int // userID → open order count (updated by service layer)
}

func NewMaxOpenOrdersRule(max int) *MaxOpenOrdersRule {
	return &MaxOpenOrdersRule{max: max, counts: make(map[string]int)}
}
func (r *MaxOpenOrdersRule) Name() string { return "MaxOpenOrders" }

func (r *MaxOpenOrdersRule) Check(_ context.Context, order domain.Order, _ domain.PortfolioSummary) error {
	r.mu.RLock()
	count := r.counts[order.UserID.String()]
	r.mu.RUnlock()
	if count >= r.max {
		return fmt.Errorf("%w: %d open orders (max %d)", domain.ErrMaxOpenOrders, count, r.max)
	}
	return nil
}

// SetCount allows the service layer to update the open order count for a user.
func (r *MaxOpenOrdersRule) SetCount(userID string, count int) {
	r.mu.Lock()
	r.counts[userID] = count
	r.mu.Unlock()
}

// DailyLossRule rejects orders when today's realised loss exceeds the limit.
type DailyLossRule struct {
	limit float64 // positive number, e.g. 1000 means "stop at -$1000"
}

func NewDailyLossRule(limit float64) *DailyLossRule { return &DailyLossRule{limit} }
func (r *DailyLossRule) Name() string               { return "DailyLoss" }

func (r *DailyLossRule) Check(_ context.Context, _ domain.Order, portfolio domain.PortfolioSummary) error {
	// DailyLoss is stored as a negative number when there's a loss
	if -portfolio.DailyLoss >= r.limit {
		return fmt.Errorf("%w: daily loss %.2f exceeds limit %.2f",
			domain.ErrDailyLossExceeded, -portfolio.DailyLoss, r.limit)
	}
	return nil
}

// DuplicateTradeWindowRule rejects orders for the same symbol within a time window.
type DuplicateTradeWindowRule struct {
	window time.Duration

	mu    sync.RWMutex
	last  map[string]time.Time // "userID:symbol" → last order time
}

func NewDuplicateTradeWindowRule(window time.Duration) *DuplicateTradeWindowRule {
	return &DuplicateTradeWindowRule{
		window: window,
		last:   make(map[string]time.Time),
	}
}
func (r *DuplicateTradeWindowRule) Name() string { return "DuplicateTradeWindow" }

func (r *DuplicateTradeWindowRule) Check(_ context.Context, order domain.Order, _ domain.PortfolioSummary) error {
	key := order.UserID.String() + ":" + order.Symbol
	r.mu.RLock()
	last, seen := r.last[key]
	r.mu.RUnlock()
	if seen && time.Since(last) < r.window {
		return fmt.Errorf("%w: last trade for %s was %s ago (window %s)",
			domain.ErrDuplicateOrder, order.Symbol, time.Since(last).Round(time.Second), r.window)
	}
	r.mu.Lock()
	r.last[key] = time.Now()
	r.mu.Unlock()
	return nil
}

// MaxPortfolioExposureRule rejects if total portfolio market value exceeds limit.
type MaxPortfolioExposureRule struct {
	limit float64
}

func NewMaxPortfolioExposureRule(limit float64) *MaxPortfolioExposureRule {
	return &MaxPortfolioExposureRule{limit}
}
func (r *MaxPortfolioExposureRule) Name() string { return "MaxPortfolioExposure" }

func (r *MaxPortfolioExposureRule) Check(_ context.Context, _ domain.Order, portfolio domain.PortfolioSummary) error {
	if portfolio.TotalValue > r.limit {
		return fmt.Errorf("%w: portfolio value %.2f > limit %.2f",
			domain.ErrPortfolioExposure, portfolio.TotalValue, r.limit)
	}
	return nil
}

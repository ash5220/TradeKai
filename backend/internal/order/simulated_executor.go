package order

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// SimulatedExecutor fills orders instantly with configurable latency and slippage.
// It implements domain.OrderExecutor.
type SimulatedExecutor struct {
	mu       sync.RWMutex
	statuses map[string]domain.OrderStatus
	fills    map[string]float64 // orderID → fill price

	latency  time.Duration
	slippage float64 // fractional slippage, e.g. 0.001 = 0.1%
}

// NewSimulatedExecutor creates a SimulatedExecutor.
func NewSimulatedExecutor(latency time.Duration, slippage float64) *SimulatedExecutor {
	return &SimulatedExecutor{
		statuses: make(map[string]domain.OrderStatus),
		fills:    make(map[string]float64),
		latency:  latency,
		slippage: slippage,
	}
}

// PlaceOrder records the order as filled after the configured latency.
func (e *SimulatedExecutor) PlaceOrder(_ context.Context, order domain.Order) (string, error) {
	exchangeID := fmt.Sprintf("sim-%d", time.Now().UnixNano())

	// Simulate fill with slippage
	slip := 1 + (rand.Float64()*2-1)*e.slippage
	fillPrice := order.LimitPrice
	if fillPrice == nil {
		p := 100.0 // placeholder; real use would require a price source
		fillPrice = &p
	}
	fill := *fillPrice * slip

	e.mu.Lock()
	e.statuses[exchangeID] = domain.OrderStatusSubmitted
	e.fills[exchangeID] = fill
	e.mu.Unlock()

	// Async fill after latency
	go func() {
		time.Sleep(e.latency)
		e.mu.Lock()
		e.statuses[exchangeID] = domain.OrderStatusFilled
		e.mu.Unlock()
	}()

	return exchangeID, nil
}

// CancelOrder cancels a simulated order if it has not yet been filled.
func (e *SimulatedExecutor) CancelOrder(_ context.Context, orderID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	status, ok := e.statuses[orderID]
	if !ok {
		return fmt.Errorf("%w: %s", domain.ErrNotFound, orderID)
	}
	if status == domain.OrderStatusFilled {
		return domain.ErrOrderAlreadyTerminal
	}
	e.statuses[orderID] = domain.OrderStatusCancelled
	return nil
}

// GetOrderStatus returns the current simulated order status.
func (e *SimulatedExecutor) GetOrderStatus(_ context.Context, orderID string) (domain.OrderStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	status, ok := e.statuses[orderID]
	if !ok {
		return "", fmt.Errorf("%w: %s", domain.ErrNotFound, orderID)
	}
	return status, nil
}

// Package domain contains pure business entities and port interfaces.
// This package has ZERO external dependencies — only the Go standard library.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderSide represents the direction of an order.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderType represents the execution constraint for an order.
type OrderType string

const (
	OrderTypeMarket OrderType = "market"
	OrderTypeLimit  OrderType = "limit"
	OrderTypeStop   OrderType = "stop"
)

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusSubmitted       OrderStatus = "submitted"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCancelled       OrderStatus = "cancelled"
	OrderStatusRejected        OrderStatus = "rejected"
)

// Order represents a trade instruction in the system.
type Order struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Symbol         string
	Side           OrderSide
	Type           OrderType
	Qty            float64
	LimitPrice     *float64   // nil for market orders
	FilledQty      float64
	FilledAvgPrice *float64
	Status         OrderStatus
	ExchangeID     string    // ID returned by the exchange after submission
	IdempotencyKey string    // user_id + symbol + signal_id dedup key
	CreatedAt      time.Time
	UpdatedAt      time.Time
	FilledAt       *time.Time
}

// IsBuyOrder returns true when the order is a buy.
func (o *Order) IsBuyOrder() bool { return o.Side == OrderSideBuy }

// IsTerminal returns true when the order is in a final state.
func (o *Order) IsTerminal() bool {
	return o.Status == OrderStatusFilled ||
		o.Status == OrderStatusCancelled ||
		o.Status == OrderStatusRejected
}

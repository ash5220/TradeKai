package order

import (
	"context"
	"fmt"

	alpaca "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// AlpacaExecutor submits orders to Alpaca paper trading.
// It implements domain.OrderExecutor.
type AlpacaExecutor struct {
	client *alpaca.Client
}

// NewAlpacaExecutor creates an AlpacaExecutor using the provided credentials.
func NewAlpacaExecutor(apiKey, apiSecret, baseURL string) *AlpacaExecutor {
	client := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   baseURL,
	})
	return &AlpacaExecutor{client: client}
}

// PlaceOrder submits the order to Alpaca and returns the exchange order ID.
func (e *AlpacaExecutor) PlaceOrder(_ context.Context, order domain.Order) (string, error) {
	qty := decimal.NewFromFloat(order.Qty)

	req := alpaca.PlaceOrderRequest{
		Symbol:      order.Symbol,
		Qty:         &qty,
		Side:        alpaca.Side(order.Side),
		Type:        alpaca.OrderType(order.Type),
		TimeInForce: alpaca.Day,
	}
	if order.LimitPrice != nil {
		lp := decimal.NewFromFloat(*order.LimitPrice)
		req.LimitPrice = &lp
	}

	placed, err := e.client.PlaceOrder(req)
	if err != nil {
		return "", fmt.Errorf("alpaca place order: %w", err)
	}
	return placed.ID, nil
}

// CancelOrder cancels the given order at Alpaca.
func (e *AlpacaExecutor) CancelOrder(_ context.Context, orderID string) error {
	if err := e.client.CancelOrder(orderID); err != nil {
		return fmt.Errorf("alpaca cancel order: %w", err)
	}
	return nil
}

// GetOrderStatus polls the order status from Alpaca.
func (e *AlpacaExecutor) GetOrderStatus(_ context.Context, orderID string) (domain.OrderStatus, error) {
	o, err := e.client.GetOrder(orderID)
	if err != nil {
		return "", fmt.Errorf("alpaca get order: %w", err)
	}
	return domain.OrderStatus(o.Status), nil
}

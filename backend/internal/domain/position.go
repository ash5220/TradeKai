package domain

import "github.com/google/uuid"

// Position represents the current holdings for a user in a single symbol.
type Position struct {
	UserID         uuid.UUID
	Symbol         string
	Qty            float64
	AvgPrice       float64
	CurrentPrice   float64
	UnrealizedPnL  float64
	RealizedPnL    float64
}

// UpdateCurrentPrice recalculates UnrealizedPnL from the latest market price.
func (p *Position) UpdateCurrentPrice(price float64) {
	p.CurrentPrice = price
	p.UnrealizedPnL = (price - p.AvgPrice) * p.Qty
}

// TotalPnL returns the sum of realised and unrealised PnL.
func (p *Position) TotalPnL() float64 {
	return p.RealizedPnL + p.UnrealizedPnL
}

// MarketValue returns the current notional value of the position.
func (p *Position) MarketValue() float64 {
	return p.CurrentPrice * p.Qty
}

// PortfolioSummary aggregates all open positions for a user.
type PortfolioSummary struct {
	UserID        uuid.UUID
	Positions     []Position
	TotalValue    float64
	TotalPnL      float64
	DailyLoss     float64 // running realised loss today (negative = loss)
}

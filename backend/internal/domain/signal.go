package domain

import (
	"time"

	"github.com/google/uuid"
)

// SignalType classifies the action recommended by a strategy.
type SignalType string

const (
	SignalBuy  SignalType = "buy"
	SignalSell SignalType = "sell"
	SignalHold SignalType = "hold"
)

// TradeSignal is the output of a strategy evaluation.
type TradeSignal struct {
	ID         uuid.UUID
	Strategy   string
	Symbol     string
	Type       SignalType
	Confidence float64   // 0–1 where 1 is highest confidence
	Price      float64   // price at signal generation time
	Timestamp  time.Time
}

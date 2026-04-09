package strategy

import (
	"time"

	"github.com/google/uuid"
	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/strategy/indicator"
)

// RSIStrategy generates buy/sell signals based on RSI thresholds.
// Buy when RSI < oversoldThreshold; Sell when RSI > overboughtThreshold.
type RSIStrategy struct {
	period             int
	oversoldThreshold  float64
	overboughtThreshold float64
}

// NewRSIStrategy creates an RSI-based strategy with configurable parameters.
func NewRSIStrategy(period int, oversold, overbought float64) *RSIStrategy {
	return &RSIStrategy{
		period:             period,
		oversoldThreshold:  oversold,
		overboughtThreshold: overbought,
	}
}

// Name implements domain.Strategy.
func (s *RSIStrategy) Name() string { return "RSI" }

// RequiredIndicators implements domain.Strategy.
func (s *RSIStrategy) RequiredIndicators() []domain.Indicator {
	return []domain.Indicator{indicator.NewRSI(s.period)}
}

// Evaluate implements domain.Strategy.
func (s *RSIStrategy) Evaluate(candle domain.Candle, indicators map[string]float64) domain.TradeSignal {
	rsiKey := indicator.NewRSI(s.period).Name()
	rsi, ok := indicators[rsiKey]

	sig := domain.TradeSignal{
		ID:        uuid.New(),
		Strategy:  s.Name(),
		Symbol:    candle.Symbol,
		Type:      domain.SignalHold,
		Price:     candle.Close,
		Timestamp: time.Now(),
	}

	if !ok {
		return sig
	}

	switch {
	case rsi < s.oversoldThreshold:
		sig.Type = domain.SignalBuy
		// Confidence scales with how far below the threshold RSI is
		sig.Confidence = (s.oversoldThreshold - rsi) / s.oversoldThreshold
	case rsi > s.overboughtThreshold:
		sig.Type = domain.SignalSell
		sig.Confidence = (rsi - s.overboughtThreshold) / (100 - s.overboughtThreshold)
	}

	return sig
}

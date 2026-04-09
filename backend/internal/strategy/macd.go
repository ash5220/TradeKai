package strategy

import (
	"time"

	"github.com/google/uuid"
	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/strategy/indicator"
)

// MACDCrossoverStrategy generates signals on MACD line / signal line crossovers.
type MACDCrossoverStrategy struct {
	fast   int
	slow   int
	signal int

	prevMACD   float64
	prevSignal float64
	first      bool
}

// NewMACDCrossoverStrategy creates a MACD crossover strategy.
func NewMACDCrossoverStrategy(fast, slow, signal int) *MACDCrossoverStrategy {
	return &MACDCrossoverStrategy{fast: fast, slow: slow, signal: signal, first: true}
}

// Name implements domain.Strategy.
func (s *MACDCrossoverStrategy) Name() string { return "MACD" }

// RequiredIndicators implements domain.Strategy.
func (s *MACDCrossoverStrategy) RequiredIndicators() []domain.Indicator {
	return []domain.Indicator{indicator.NewMACD(s.fast, s.slow, s.signal)}
}

// Evaluate implements domain.Strategy.
func (s *MACDCrossoverStrategy) Evaluate(candle domain.Candle, indicators map[string]float64) domain.TradeSignal {
	macdInd := indicator.NewMACD(s.fast, s.slow, s.signal)
	macdKey := macdInd.Name()
	signalKey := macdKey + "_signal"

	sig := domain.TradeSignal{
		ID:        uuid.New(),
		Strategy:  s.Name(),
		Symbol:    candle.Symbol,
		Type:      domain.SignalHold,
		Price:     candle.Close,
		Timestamp: time.Now(),
	}

	macdVal, macdOk := indicators[macdKey]
	signalVal, signalOk := indicators[signalKey]

	if !macdOk || !signalOk {
		return sig
	}

	if s.first {
		s.prevMACD = macdVal
		s.prevSignal = signalVal
		s.first = false
		return sig
	}

	// Bullish crossover: MACD crosses above signal line
	if s.prevMACD <= s.prevSignal && macdVal > signalVal {
		sig.Type = domain.SignalBuy
		sig.Confidence = 0.7
	}
	// Bearish crossover: MACD crosses below signal line
	if s.prevMACD >= s.prevSignal && macdVal < signalVal {
		sig.Type = domain.SignalSell
		sig.Confidence = 0.7
	}

	s.prevMACD = macdVal
	s.prevSignal = signalVal

	return sig
}

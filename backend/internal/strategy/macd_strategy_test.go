package strategy_test

import (
	"testing"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/strategy"
	"github.com/rashevskyv/tradekai/internal/strategy/indicator"
)

func TestMACDCrossoverStrategyEvaluate(t *testing.T) {
	t.Helper()
	s := strategy.NewMACDCrossoverStrategy(12, 26, 9)
	key := indicator.NewMACD(12, 26, 9).Name()
	sigKey := key + "_signal"
	candle := domain.Candle{Symbol: "AAPL", Close: 101}

	first := s.Evaluate(candle, map[string]float64{key: -1.0, sigKey: -0.5})
	if first.Type != domain.SignalHold {
		t.Errorf("Evaluate(first) = %q, want %q", first.Type, domain.SignalHold)
	}

	bullish := s.Evaluate(candle, map[string]float64{key: 0.25, sigKey: 0.1})
	if bullish.Type != domain.SignalBuy {
		t.Errorf("Evaluate(bullish crossover) = %q, want %q", bullish.Type, domain.SignalBuy)
	}

	bearish := s.Evaluate(candle, map[string]float64{key: -0.2, sigKey: 0.05})
	if bearish.Type != domain.SignalSell {
		t.Errorf("Evaluate(bearish crossover) = %q, want %q", bearish.Type, domain.SignalSell)
	}

	missing := s.Evaluate(candle, map[string]float64{key: 1.0})
	if missing.Type != domain.SignalHold {
		t.Errorf("Evaluate(missing signal line) = %q, want %q", missing.Type, domain.SignalHold)
	}
}

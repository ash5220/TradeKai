package strategy_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/strategy"
	"github.com/rashevskyv/tradekai/internal/strategy/indicator"
)

func TestRSIStrategyEvaluate(t *testing.T) {
	t.Helper()
	s := strategy.NewRSIStrategy(14, 30, 70)
	rsiKey := indicator.NewRSI(14).Name()
	candle := domain.Candle{Symbol: "AAPL", Close: 100}

	tests := []struct {
		name       string
		indicators map[string]float64
		wantType   domain.SignalType
	}{
		{
			name:       "hold when rsi missing",
			indicators: map[string]float64{},
			wantType:   domain.SignalHold,
		},
		{
			name:       "buy when oversold",
			indicators: map[string]float64{rsiKey: 20},
			wantType:   domain.SignalBuy,
		},
		{
			name:       "sell when overbought",
			indicators: map[string]float64{rsiKey: 80},
			wantType:   domain.SignalSell,
		},
		{
			name:       "hold in neutral band",
			indicators: map[string]float64{rsiKey: 50},
			wantType:   domain.SignalHold,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := s.Evaluate(candle, tc.indicators)
			if diff := cmp.Diff(tc.wantType, got.Type); diff != "" {
				t.Errorf("RSIStrategy.Evaluate() signal type mismatch (-want +got):\n%s", diff)
			}
			if got.Symbol != candle.Symbol {
				t.Errorf("RSIStrategy.Evaluate() symbol = %q, want %q", got.Symbol, candle.Symbol)
			}
		})
	}
}

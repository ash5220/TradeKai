package indicator_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/strategy/indicator"
)

func candle(close float64) domain.Candle {
	return domain.Candle{Close: close}
}

func TestSMA(t *testing.T) {
	t.Helper()
	tests := []struct {
		name    string
		period  int
		prices  []float64
		want    []float64 // expected SMA at each step (0 = not ready)
		wantOk  []bool
	}{
		{
			name:   "period 3 basic",
			period: 3,
			prices: []float64{10, 20, 30, 40, 50},
			want:   []float64{0, 0, 20, 30, 40},
			wantOk: []bool{false, false, true, true, true},
		},
		{
			name:   "period 1 returns each value",
			period: 1,
			prices: []float64{5, 10, 15},
			want:   []float64{5, 10, 15},
			wantOk: []bool{true, true, true},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sma := indicator.NewSMA(tc.period)
			for i, price := range tc.prices {
				got, ok := sma.Add(candle(price))
				if diff := cmp.Diff(tc.wantOk[i], ok); diff != "" {
					t.Errorf("SMA.Add(%v) ok mismatch at step %d (-want +got):\n%s", price, i, diff)
				}
				if ok {
					if diff := cmp.Diff(tc.want[i], got); diff != "" {
						t.Errorf("SMA.Add(%v) value mismatch at step %d (-want +got):\n%s", price, i, diff)
					}
				}
			}
		})
	}
}

func TestEMA(t *testing.T) {
	t.Helper()
	// With period=3: multiplier = 2/(3+1) = 0.5
	// Seed = SMA of first 3 values = (10+20+30)/3 = 20
	// step 4: 40*0.5 + 20*0.5 = 30
	// step 5: 50*0.5 + 30*0.5 = 40
	prices := []float64{10, 20, 30, 40, 50}
	wantOk := []bool{false, false, true, true, true}
	wantVals := []float64{0, 0, 20, 30, 40}

	ema := indicator.NewEMA(3)
	for i, price := range prices {
		got, ok := ema.Add(candle(price))
		if ok != wantOk[i] {
			t.Errorf("EMA.Add(%v) ok=%v, want %v at step %d", price, ok, wantOk[i], i)
		}
		if ok && got != wantVals[i] {
			t.Errorf("EMA.Add(%v) = %v, want %v at step %d", price, got, wantVals[i], i)
		}
	}
}

func TestRSI_Overbought(t *testing.T) {
	t.Helper()
	// All gains → RSI should approach 100
	rsi := indicator.NewRSI(3)
	prices := []float64{10, 11, 12, 13, 14, 15}
	var lastRSI float64
	for _, p := range prices {
		v, ok := rsi.Add(candle(p))
		if ok {
			lastRSI = v
		}
	}
	if lastRSI < 90 {
		t.Errorf("RSI with all gains = %v, want > 90", lastRSI)
	}
}

func TestRSI_Oversold(t *testing.T) {
	t.Helper()
	// All losses → RSI should approach 0
	rsi := indicator.NewRSI(3)
	prices := []float64{15, 14, 13, 12, 11, 10}
	var lastRSI float64
	for _, p := range prices {
		v, ok := rsi.Add(candle(p))
		if ok {
			lastRSI = v
		}
	}
	if lastRSI > 10 {
		t.Errorf("RSI with all losses = %v, want < 10", lastRSI)
	}
}

func TestRSI_Reset(t *testing.T) {
	t.Helper()
	rsi := indicator.NewRSI(3)
	for _, p := range []float64{10, 20, 30, 40} {
		rsi.Add(candle(p))
	}
	rsi.Reset()
	_, ok := rsi.Value()
	if ok {
		t.Error("RSI.Value() ok=true after Reset, want false")
	}
}

func TestSMA_Reset(t *testing.T) {
	t.Helper()
	sma := indicator.NewSMA(3)
	for _, p := range []float64{10, 20, 30} {
		sma.Add(candle(p))
	}
	sma.Reset()
	_, ok := sma.Value()
	if ok {
		t.Error("SMA.Value() ok=true after Reset, want false")
	}
}

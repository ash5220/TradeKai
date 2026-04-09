package indicator

import (
	"fmt"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// EMA is an Exponential Moving Average indicator.
// After `period` candles it switches from SMA seed to EMA.
type EMA struct {
	period     int
	multiplier float64
	value      float64
	count      int
	sumSeed    float64 // accumulates until period bars for the SMA seed
	ready      bool
}

// NewEMA creates an Exponential Moving Average with the given period.
func NewEMA(period int) *EMA {
	return &EMA{
		period:     period,
		multiplier: 2.0 / float64(period+1),
	}
}

// Name implements domain.Indicator.
func (e *EMA) Name() string { return fmt.Sprintf("EMA(%d)", e.period) }

// Add feeds the next candle's close price and returns the current EMA value.
func (e *EMA) Add(candle domain.Candle) (float64, bool) {
	e.count++
	if e.count < e.period {
		e.sumSeed += candle.Close
		return 0, false
	}
	if e.count == e.period {
		// Seed: use SMA of first `period` candles
		e.sumSeed += candle.Close
		e.value = e.sumSeed / float64(e.period)
		e.ready = true
		return e.value, true
	}
	// Standard EMA update
	e.value = candle.Close*e.multiplier + e.value*(1-e.multiplier)
	return e.value, true
}

// Value implements domain.Indicator.
func (e *EMA) Value() (float64, bool) {
	return e.value, e.ready
}

// Reset clears all internal state.
func (e *EMA) Reset() {
	e.value = 0
	e.count = 0
	e.sumSeed = 0
	e.ready = false
}

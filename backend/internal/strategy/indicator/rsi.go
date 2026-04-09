package indicator

import (
	"fmt"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// RSI is a Relative Strength Index indicator (Wilder's smoothing method).
type RSI struct {
	period  int
	count   int
	prevClose float64
	avgGain float64
	avgLoss float64
	value   float64
	ready   bool
}

// NewRSI creates an RSI with the given period (typically 14).
func NewRSI(period int) *RSI {
	return &RSI{period: period}
}

// Name implements domain.Indicator.
func (r *RSI) Name() string { return fmt.Sprintf("RSI(%d)", r.period) }

// Add feeds the next candle's close price and returns the current RSI value.
// RSI formula: RS = avgGain / avgLoss; RSI = 100 - 100/(1+RS)
func (r *RSI) Add(candle domain.Candle) (float64, bool) {
	r.count++

	if r.count == 1 {
		r.prevClose = candle.Close
		return 0, false
	}

	change := candle.Close - r.prevClose
	gain := max(change, 0)
	loss := max(-change, 0)
	r.prevClose = candle.Close

	if r.count <= r.period {
		// Accumulate for the initial SMA seed
		r.avgGain += gain
		r.avgLoss += loss
		if r.count == r.period {
			r.avgGain /= float64(r.period)
			r.avgLoss /= float64(r.period)
			r.value = r.rsiFromAvg()
			r.ready = true
			return r.value, true
		}
		return 0, false
	}

	// Wilder's smoothing
	r.avgGain = (r.avgGain*float64(r.period-1) + gain) / float64(r.period)
	r.avgLoss = (r.avgLoss*float64(r.period-1) + loss) / float64(r.period)
	r.value = r.rsiFromAvg()
	return r.value, true
}

func (r *RSI) rsiFromAvg() float64 {
	if r.avgLoss == 0 {
		return 100
	}
	rs := r.avgGain / r.avgLoss
	return 100 - (100 / (1 + rs))
}

// Value implements domain.Indicator.
func (r *RSI) Value() (float64, bool) {
	return r.value, r.ready
}

// Reset clears all internal state.
func (r *RSI) Reset() {
	r.count = 0
	r.prevClose = 0
	r.avgGain = 0
	r.avgLoss = 0
	r.value = 0
	r.ready = false
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

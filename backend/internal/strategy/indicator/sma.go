// Package indicator contains streaming technical indicator implementations.
// All indicators implement domain.Indicator (feed one candle at a time).
package indicator

import (
	"fmt"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// SMA is a Simple Moving Average indicator.
type SMA struct {
	period int
	buf    []float64 // circular buffer
	pos    int
	count  int
	sum    float64
}

// NewSMA creates a Simple Moving Average with the given period.
func NewSMA(period int) *SMA {
	return &SMA{period: period, buf: make([]float64, period)}
}

// Name implements domain.Indicator.
func (s *SMA) Name() string { return fmt.Sprintf("SMA(%d)", s.period) }

// Add feeds the next candle's close price and returns the current SMA value.
func (s *SMA) Add(candle domain.Candle) (float64, bool) {
	if s.count == s.period {
		// Remove the oldest value that is being overwritten
		s.sum -= s.buf[s.pos]
	} else {
		s.count++
	}
	s.buf[s.pos] = candle.Close
	s.sum += candle.Close
	s.pos = (s.pos + 1) % s.period

	if s.count < s.period {
		return 0, false
	}
	return s.sum / float64(s.period), true
}

// Value implements domain.Indicator.
func (s *SMA) Value() (float64, bool) {
	if s.count < s.period {
		return 0, false
	}
	return s.sum / float64(s.period), true
}

// Reset clears all internal state.
func (s *SMA) Reset() {
	s.buf = make([]float64, s.period)
	s.pos = 0
	s.count = 0
	s.sum = 0
}

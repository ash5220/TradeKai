package indicator

import (
	"fmt"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// MACD is the Moving Average Convergence Divergence indicator.
// It outputs the MACD line value; the signal line and histogram are also tracked.
type MACD struct {
	fast   *EMA
	slow   *EMA
	signal *EMA // signal line EMA applied to MACD line values

	macdValue   float64
	signalValue float64
	histogram   float64
	ready       bool

	// synthetic candle used to feed signal EMA with MACD values
	macdBuf []float64
}

// NewMACD creates a MACD indicator (typically fast=12, slow=26, signal=9).
func NewMACD(fast, slow, signal int) *MACD {
	return &MACD{
		fast:   NewEMA(fast),
		slow:   NewEMA(slow),
		signal: NewEMA(signal),
	}
}

// Name implements domain.Indicator.
func (m *MACD) Name() string {
	return fmt.Sprintf("MACD(%d,%d,%d)", m.fast.period, m.slow.period, m.signal.period)
}

// Add feeds the next candle and returns the MACD line value when ready.
// MACD is ready only once the slow EMA and signal EMA both have enough data.
func (m *MACD) Add(candle domain.Candle) (float64, bool) {
	fastVal, fastReady := m.fast.Add(candle)
	slowVal, slowReady := m.slow.Add(candle)

	if !fastReady || !slowReady {
		return 0, false
	}

	macdLine := fastVal - slowVal

	// Feed the MACD line value into the signal EMA (wrapped as a synthetic candle)
	synth := domain.Candle{Close: macdLine}
	sigVal, sigReady := m.signal.Add(synth)

	if !sigReady {
		return 0, false
	}

	m.macdValue = macdLine
	m.signalValue = sigVal
	m.histogram = macdLine - sigVal
	m.ready = true

	return m.macdValue, true
}

// SignalLine returns the current signal line value.
func (m *MACD) SignalLine() float64 { return m.signalValue }

// Histogram returns the current histogram value (MACD - signal).
func (m *MACD) Histogram() float64 { return m.histogram }

// Value implements domain.Indicator.
func (m *MACD) Value() (float64, bool) {
	return m.macdValue, m.ready
}

// Reset clears all internal state.
func (m *MACD) Reset() {
	m.fast.Reset()
	m.slow.Reset()
	m.signal.Reset()
	m.macdValue = 0
	m.signalValue = 0
	m.histogram = 0
	m.ready = false
}

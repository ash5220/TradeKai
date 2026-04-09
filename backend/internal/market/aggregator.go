package market

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// candleState tracks in-progress candle building for one (symbol, interval) pair.
type candleState struct {
	candle    domain.Candle
	tickCount int
	endTime   time.Time
}

// CandleSink is a function that receives a completed candle.
type CandleSink func(candle domain.Candle)

// Aggregator converts a stream of Ticks into OHLCV Candles at a fixed interval.
// Call Subscribe to receive ticks, then read completed candles from the sink.
type Aggregator struct {
	interval time.Duration
	sink     CandleSink

	mu     sync.Mutex
	states map[string]*candleState // symbol → in-progress candle
}

// NewAggregator creates an Aggregator that emits candles at the given interval
// and calls sink for each completed candle.
func NewAggregator(interval time.Duration, sink CandleSink) *Aggregator {
	return &Aggregator{
		interval: interval,
		sink:     sink,
		states:   make(map[string]*candleState),
	}
}

// Process feeds a tick into the aggregator.
func (a *Aggregator) Process(tick domain.Tick) {
	a.mu.Lock()
	defer a.mu.Unlock()

	state, ok := a.states[tick.Symbol]
	if !ok || tick.Timestamp.After(state.endTime) {
		if ok && state.tickCount > 0 {
			// Flush the completed candle before starting a new one
			completed := state.candle
			go a.sink(completed)
		}
		// Align the candle start to the interval boundary
		start := tick.Timestamp.Truncate(a.interval)
		state = &candleState{
			candle: domain.Candle{
				Symbol:    tick.Symbol,
				Open:      tick.Price,
				High:      tick.Price,
				Low:       tick.Price,
				Close:     tick.Price,
				Volume:    tick.Volume,
				Interval:  a.interval,
				Timestamp: start,
			},
			tickCount: 1,
			endTime:   start.Add(a.interval),
		}
		a.states[tick.Symbol] = state
		return
	}

	// Update the in-progress candle
	if tick.Price > state.candle.High {
		state.candle.High = tick.Price
	}
	if tick.Price < state.candle.Low {
		state.candle.Low = tick.Price
	}
	state.candle.Close = tick.Price
	state.candle.Volume += tick.Volume
	state.tickCount++
}

// RunFromChannel reads ticks from src until ctx is cancelled or src is closed,
// then processes each tick through the aggregator.
func (a *Aggregator) RunFromChannel(ctx context.Context, symbol string, src <-chan domain.Tick) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case tick, ok := <-src:
			if !ok {
				return fmt.Errorf("aggregator: tick channel for %s closed", symbol)
			}
			a.Process(tick)
		}
	}
}

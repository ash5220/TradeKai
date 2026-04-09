package market

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// SimulatedProvider generates random-walk tick data for development and testing.
// It implements domain.MarketDataProvider.
type SimulatedProvider struct {
	mu          sync.RWMutex
	symbols     []string
	channels    map[string]chan domain.Tick
	prices      map[string]float64 // last price per symbol
	volatility  float64            // fractional daily vol, e.g. 0.02 = 2%
	tickInterval time.Duration
	cancel      context.CancelFunc
}

// NewSimulatedProvider creates a SimulatedProvider.
// volatility is expressed as a fraction per tick (e.g. 0.001).
func NewSimulatedProvider(volatility float64, tickInterval time.Duration) *SimulatedProvider {
	return &SimulatedProvider{
		channels:    make(map[string]chan domain.Tick),
		prices:      make(map[string]float64),
		volatility:  volatility,
		tickInterval: tickInterval,
	}
}

// Connect initialises channels for all requested symbols and starts generating ticks.
func (p *SimulatedProvider) Connect(ctx context.Context, symbols []string) error {
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	p.mu.Lock()
	p.symbols = symbols
	for _, sym := range symbols {
		if _, exists := p.channels[sym]; !exists {
			// Start each symbol at a plausible price between 50 and 500
			p.prices[sym] = 50 + rand.Float64()*450
			p.channels[sym] = make(chan domain.Tick, 1024)
		}
	}
	p.mu.Unlock()

	for _, sym := range symbols {
		sym := sym
		go p.generate(ctx, sym)
	}

	return nil
}

// Subscribe returns the tick channel for a symbol. Connect must be called first.
func (p *SimulatedProvider) Subscribe(symbol string) (<-chan domain.Tick, error) {
	p.mu.RLock()
	ch, ok := p.channels[symbol]
	p.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("simulated provider: symbol %q not connected", symbol)
	}
	return ch, nil
}

// Close cancels tick generation and closes all channels.
func (p *SimulatedProvider) Close() error {
	if p.cancel != nil {
		p.cancel()
	}
	p.mu.Lock()
	for sym, ch := range p.channels {
		close(ch)
		delete(p.channels, sym)
	}
	p.mu.Unlock()
	return nil
}

func (p *SimulatedProvider) generate(ctx context.Context, symbol string) {
	ticker := time.NewTicker(p.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			p.mu.Lock()
			// Geometric Brownian Motion step
			drift := (rand.Float64() - 0.5) * p.volatility
			p.prices[symbol] *= (1 + drift)
			price := p.prices[symbol]
			ch := p.channels[symbol]
			p.mu.Unlock()

			tick := domain.Tick{
				Symbol:    symbol,
				Price:     price,
				Volume:    10 + rand.Float64()*1000,
				Timestamp: t,
			}

			select {
			case ch <- tick:
			default:
				// Channel full — drop oldest rather than blocking
				select {
				case <-ch:
				default:
				}
				select {
				case ch <- tick:
				default:
				}
			}
		}
	}
}

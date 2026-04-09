// Package market contains adapters implementing domain.MarketDataProvider.
package market

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/telemetry"
)

// HealthSnapshot contains lightweight exchange connectivity details used by
// system health endpoints.
type HealthSnapshot struct {
	Connected    bool      `json:"connected"`
	LastTickAt   time.Time `json:"lastTickAt,omitempty"`
	ActiveStream int       `json:"activeStreamCount"`
}

// Hub fans out ticks from a single provider to multiple per-symbol subscribers.
// Subscribers receive ticks via buffered channels; a slow consumer causes the
// oldest buffered tick to be dropped rather than blocking the producer.
type Hub struct {
	provider domain.MarketDataProvider
	log      *zap.Logger

	mu          sync.RWMutex
	subscribers map[string][]chan domain.Tick // symbol → subscriber channels
	lastTickAt  time.Time
	connected   bool
	streams     int

	bufSize int // per-subscriber channel capacity
}

// NewHub creates a Hub that reads from provider.
func NewHub(provider domain.MarketDataProvider, log *zap.Logger) *Hub {
	return &Hub{
		provider:    provider,
		log:         log,
		subscribers: make(map[string][]chan domain.Tick),
		bufSize:     1024,
	}
}

// Start connects the underlying provider for the given symbols and begins
// fanning out ticks to subscribers. It blocks until ctx is cancelled.
func (h *Hub) Start(ctx context.Context, symbols []string) error {
	if err := h.provider.Connect(ctx, symbols); err != nil {
		return fmt.Errorf("hub: connect provider: %w", err)
	}

	h.mu.Lock()
	h.connected = true
	h.streams = len(symbols)
	h.mu.Unlock()

	for _, sym := range symbols {
		sym := sym
		ch, err := h.provider.Subscribe(sym)
		if err != nil {
			h.mu.Lock()
			h.connected = false
			h.mu.Unlock()
			return fmt.Errorf("hub: subscribe %s: %w", sym, err)
		}
		go h.fanOut(ctx, sym, ch)
	}

	<-ctx.Done()

	if err := h.provider.Close(); err != nil {
		h.log.Warn("hub: close provider", zap.Error(err))
	}

	h.mu.Lock()
	h.connected = false
	h.mu.Unlock()

	return nil
}

// Subscribe registers a new subscriber for symbol and returns its receive channel.
// The returned channel must not be consumed after the Hub is stopped.
func (h *Hub) Subscribe(symbol string) <-chan domain.Tick {
	ch := make(chan domain.Tick, h.bufSize)
	h.mu.Lock()
	h.subscribers[symbol] = append(h.subscribers[symbol], ch)
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes the channel from the subscriber list and closes it.
func (h *Hub) Unsubscribe(symbol string, ch <-chan domain.Tick) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.subscribers[symbol]
	for i, s := range subs {
		if s == ch {
			h.subscribers[symbol] = append(subs[:i], subs[i+1:]...)
			close(s)
			return
		}
	}
}

func (h *Hub) fanOut(ctx context.Context, symbol string, src <-chan domain.Tick) {
	for {
		select {
		case <-ctx.Done():
			return
		case tick, ok := <-src:
			if !ok {
				return
			}
			h.dispatch(symbol, tick)
		}
	}
}

func (h *Hub) dispatch(symbol string, tick domain.Tick) {
	telemetry.IncMarketDataTick(tick.Symbol)

	h.mu.Lock()
	h.lastTickAt = tick.Timestamp
	h.mu.Unlock()

	h.mu.RLock()
	subs := make([]chan domain.Tick, len(h.subscribers[symbol]))
	copy(subs, h.subscribers[symbol])
	h.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- tick:
		default:
			// drop oldest to keep producers unblocked
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

// HealthSnapshot returns exchange stream health with a simple stale-tick check.
func (h *Hub) HealthSnapshot(staleAfter time.Duration) HealthSnapshot {
	h.mu.RLock()
	connected := h.connected
	lastTickAt := h.lastTickAt
	streams := h.streams
	h.mu.RUnlock()

	if connected && !lastTickAt.IsZero() && staleAfter > 0 && time.Since(lastTickAt) > staleAfter {
		connected = false
	}

	return HealthSnapshot{
		Connected:    connected,
		LastTickAt:   lastTickAt,
		ActiveStream: streams,
	}
}

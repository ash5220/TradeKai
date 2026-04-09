package market

import (
	"context"
	"fmt"
	"sync"

	alpaca "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
	"go.uber.org/zap"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// AlpacaProvider implements domain.MarketDataProvider using Alpaca's WebSocket stream.
type AlpacaProvider struct {
	apiKey    string
	apiSecret string
	dataFeed  string
	log       *zap.Logger

	mu       sync.RWMutex
	client   *alpaca.StocksClient
	channels map[string]chan domain.Tick
}

// NewAlpacaProvider creates an AlpacaProvider.
func NewAlpacaProvider(apiKey, apiSecret, dataFeed string, log *zap.Logger) *AlpacaProvider {
	return &AlpacaProvider{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		dataFeed:  dataFeed,
		log:       log,
		channels:  make(map[string]chan domain.Tick),
	}
}

// Connect establishes the Alpaca WebSocket connection and subscribes to trades.
func (p *AlpacaProvider) Connect(ctx context.Context, symbols []string) error {
	feed := alpaca.US
	if p.dataFeed == "sip" {
		feed = alpaca.SIP
	}

	client := alpaca.NewStocksClient(feed,
		alpaca.WithCredentials(p.apiKey, p.apiSecret),
	)

	p.mu.Lock()
	p.client = client
	for _, sym := range symbols {
		p.channels[sym] = make(chan domain.Tick, 1024)
	}
	p.mu.Unlock()

	handler := func(t alpaca.Trade) {
		tick := domain.Tick{
			Symbol:    t.Symbol,
			Price:     t.Price,
			Volume:    float64(t.Size),
			Timestamp: t.Timestamp,
		}
		p.mu.RLock()
		ch, ok := p.channels[tick.Symbol]
		p.mu.RUnlock()
		if !ok {
			return
		}
		select {
		case ch <- tick:
		default:
			// drop oldest, then try again
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

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("alpaca: connect: %w", err)
	}

	if err := client.SubscribeToTrades(handler, symbols...); err != nil {
		return fmt.Errorf("alpaca: subscribe to trades: %w", err)
	}

	go func() {
		if err := client.Run(); err != nil {
			p.log.Error("alpaca stream error", zap.Error(err))
		}
	}()

	return nil
}

// Subscribe returns the tick channel for the given symbol.
func (p *AlpacaProvider) Subscribe(symbol string) (<-chan domain.Tick, error) {
	p.mu.RLock()
	ch, ok := p.channels[symbol]
	p.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("alpaca provider: symbol %q not subscribed", symbol)
	}
	return ch, nil
}

// Close shuts down the Alpaca stream and closes all channels.
func (p *AlpacaProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for sym, ch := range p.channels {
		close(ch)
		delete(p.channels, sym)
	}
	return nil
}

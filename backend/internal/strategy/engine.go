// Package strategy contains the strategy engine and built-in strategies.
package strategy

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/market"
)

// symbolWorker holds per-symbol indicator state for a running strategy.
type symbolWorker struct {
	strategy   domain.Strategy
	indicators map[string]domain.Indicator // keyed by Indicator.Name()
	candleCh   <-chan domain.Candle        // from aggregator
}

// Engine subscribes to candle events from the market hub, runs configured strategies,
// and emits TradeSignals on its output channel.
type Engine struct {
	hub     *market.Hub
	log     *zap.Logger
	signals chan domain.TradeSignal

	mu       sync.RWMutex
	workers  map[string]map[string]*symbolWorker // strategyName → symbol → worker
	cancels  map[string]context.CancelFunc        // strategyName → cancel fn
}

// NewEngine creates a strategy Engine.
func NewEngine(hub *market.Hub, log *zap.Logger) *Engine {
	return &Engine{
		hub:     hub,
		log:     log,
		signals: make(chan domain.TradeSignal, 256),
		workers: make(map[string]map[string]*symbolWorker),
		cancels: make(map[string]context.CancelFunc),
	}
}

// Signals returns the read-only channel that receives generated trade signals.
func (e *Engine) Signals() <-chan domain.TradeSignal { return e.signals }

// Start activates a strategy for a set of symbols.
// It subscribes to the market hub for each symbol and spawns a worker goroutine.
func (e *Engine) Start(ctx context.Context, strat domain.Strategy, symbols []string) error {
	name := strat.Name()

	e.mu.Lock()
	if _, running := e.cancels[name]; running {
		e.mu.Unlock()
		return domain.ErrStrategyAlreadyRunning
	}

	stratCtx, cancel := context.WithCancel(ctx)
	e.cancels[name] = cancel
	e.workers[name] = make(map[string]*symbolWorker)
	e.mu.Unlock()

	g, gCtx := errgroup.WithContext(stratCtx)

	for _, sym := range symbols {
		sym := sym
		candleCh := make(chan domain.Candle, 256)

		// Build fresh indicator instances for this (strategy, symbol) pair
		inds := make(map[string]domain.Indicator)
		for _, ind := range strat.RequiredIndicators() {
			inds[ind.Name()] = ind
		}

		w := &symbolWorker{
			strategy:   strat,
			indicators: inds,
			candleCh:   candleCh,
		}

		e.mu.Lock()
		e.workers[name][sym] = w
		e.mu.Unlock()

		// Subscribe to ticks and convert to candles (1m default)
		tickCh := e.hub.Subscribe(sym)

		g.Go(func() error {
			return e.runWorker(gCtx, name, sym, w, tickCh, candleCh)
		})
	}

	go func() {
		if err := g.Wait(); err != nil {
			e.log.Error("strategy engine worker error",
				zap.String("strategy", name), zap.Error(err))
		}
	}()

	e.log.Info("strategy started", zap.String("strategy", name),
		zap.Strings("symbols", symbols))
	return nil
}

// Stop halts a running strategy.
func (e *Engine) Stop(strategyName string) {
	e.mu.Lock()
	cancel, ok := e.cancels[strategyName]
	if ok {
		cancel()
		delete(e.cancels, strategyName)
		delete(e.workers, strategyName)
	}
	e.mu.Unlock()

	if ok {
		e.log.Info("strategy stopped", zap.String("strategy", strategyName))
	}
}

// RunningStrategies returns names of all currently active strategies.
func (e *Engine) RunningStrategies() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	names := make([]string, 0, len(e.cancels))
	for name := range e.cancels {
		names = append(names, name)
	}
	return names
}

func (e *Engine) runWorker(
	ctx context.Context,
	stratName, symbol string,
	w *symbolWorker,
	tickCh <-chan domain.Tick,
	candleCh chan<- domain.Candle,
) error {
	// Use a 1-minute aggregator that writes to candleCh
	agg := market.NewAggregator(time.Minute, func(c domain.Candle) {
		select {
		case candleCh <- c:
		default:
		}
	})

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return agg.RunFromChannel(gCtx, symbol, tickCh)
	})
	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return nil
			case candle, ok := <-candleCh:
				if !ok {
					return nil
				}
				e.evaluate(w, candle)
			}
		}
	})
	return g.Wait()
}

func (e *Engine) evaluate(w *symbolWorker, candle domain.Candle) {
	vals := make(map[string]float64, len(w.indicators))
	for name, ind := range w.indicators {
		v, ok := ind.Add(candle)
		if ok {
			vals[name] = v
		}
	}
	sig := w.strategy.Evaluate(candle, vals)
	if sig.Type == domain.SignalHold {
		return
	}
	select {
	case e.signals <- sig:
	default:
		e.log.Warn("signal channel full, dropping signal",
			zap.String("strategy", sig.Strategy),
			zap.String("symbol", sig.Symbol))
	}
}

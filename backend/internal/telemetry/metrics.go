package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	orderExecutionLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tradekai_order_execution_latency_seconds",
			Help:    "Latency of order execution attempts at the configured executor.",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"executor", "status"},
	)

	marketDataTicksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tradekai_market_data_ticks_total",
			Help: "Total number of market data ticks received per symbol.",
		},
		[]string{"symbol"},
	)

	activeWebSocketConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tradekai_active_websocket_connections",
			Help: "Current number of active WebSocket connections.",
		},
	)

	strategySignalsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tradekai_strategy_signals_total",
			Help: "Total number of emitted strategy signals.",
		},
		[]string{"strategy", "symbol", "signal"},
	)

	riskChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tradekai_risk_checks_total",
			Help: "Total number of risk checks by rule and result.",
		},
		[]string{"rule", "result"},
	)

	orderStatusTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tradekai_order_status_total",
			Help: "Total number of order status transitions.",
		},
		[]string{"status"},
	)
)

func ObserveOrderExecutionLatency(executor, status string, seconds float64) {
	orderExecutionLatency.WithLabelValues(executor, status).Observe(seconds)
}

func IncMarketDataTick(symbol string) {
	marketDataTicksTotal.WithLabelValues(symbol).Inc()
}

func SetActiveWebSocketConnections(count int) {
	activeWebSocketConnections.Set(float64(count))
}

func IncStrategySignal(strategyName, symbol, signalType string) {
	strategySignalsTotal.WithLabelValues(strategyName, symbol, signalType).Inc()
}

func IncRiskCheck(rule, result string) {
	riskChecksTotal.WithLabelValues(rule, result).Inc()
}

func IncOrderStatus(status string) {
	orderStatusTotal.WithLabelValues(status).Inc()
}

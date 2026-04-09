package domain

import "time"

// Tick is a single price update for a symbol.
type Tick struct {
	Symbol    string
	Price     float64
	Volume    float64
	Timestamp time.Time
}

// Candle is an OHLCV bar for a symbol over an interval.
type Candle struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Interval  time.Duration
	Timestamp time.Time // start of the interval
}

// Quote represents the current bid/ask for a symbol.
type Quote struct {
	Symbol    string
	BidPrice  float64
	AskPrice  float64
	BidSize   float64
	AskSize   float64
	Timestamp time.Time
}

// Symbol represents a tradeable instrument.
type Symbol struct {
	Ticker   string
	Name     string
	Exchange string
	Tradable bool
}

// MidPrice returns the mid-point between bid and ask.
func (q *Quote) MidPrice() float64 {
	return (q.BidPrice + q.AskPrice) / 2
}

// Spread returns the absolute bid-ask spread.
func (q *Quote) Spread() float64 {
	return q.AskPrice - q.BidPrice
}

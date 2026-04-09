package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rashevskyv/tradekai/internal/store/generated"
)

// MarketHandler handles market data endpoints.
type MarketHandler struct {
	queries *generated.Queries
}

// NewMarketHandler creates a MarketHandler.
func NewMarketHandler(db *pgxpool.Pool) *MarketHandler {
	return &MarketHandler{queries: generated.New(db)}
}

// Candles godoc
// @Summary Get historical candles for a symbol
// @Tags market
// @Produce json
// @Param symbol   path   string true  "Ticker symbol (e.g. AAPL)"
// @Param interval query  string false "Candle interval (1m, 5m, 1h)" default(1m)
// @Param from     query  string false "Start time RFC3339" default(24h ago)
// @Param to       query  string false "End time RFC3339"   default(now)
// @Param limit    query  int    false "Max candles"        default(500)
// @Success 200 {array} generated.Candle
// @Router /market/candles/{symbol} [get]
func (h *MarketHandler) Candles(c *gin.Context) {
	symbol := c.Param("symbol")
	interval := c.DefaultQuery("interval", "1m")

	now := time.Now()
	from := now.Add(-24 * time.Hour)
	to := now

	if f := c.Query("from"); f != "" {
		if t, err := time.Parse(time.RFC3339, f); err == nil {
			from = t
		}
	}
	if t := c.Query("to"); t != "" {
		if p, err := time.Parse(time.RFC3339, t); err == nil {
			to = p
		}
	}

	limit := int32(500)
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "500")); err == nil {
		limit = int32(l)
	}

	candles, err := h.queries.ListCandles(c.Request.Context(), generated.ListCandlesParams{
		Symbol:   symbol,
		Interval: interval,
		Ts:       pgTimestamp(from),
		Ts_2:     pgTimestamp(to),
		Limit:    limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch candles"})
		return
	}

	c.JSON(http.StatusOK, candles)
}

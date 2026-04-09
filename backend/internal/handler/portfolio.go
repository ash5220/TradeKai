package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rashevskyv/tradekai/internal/store/generated"
)

// PortfolioHandler handles portfolio endpoints.
type PortfolioHandler struct {
	queries *generated.Queries
}

// NewPortfolioHandler creates a PortfolioHandler.
func NewPortfolioHandler(db *pgxpool.Pool) *PortfolioHandler {
	return &PortfolioHandler{queries: generated.New(db)}
}

// Positions godoc
// @Summary List open positions for the authenticated user
// @Tags portfolio
// @Security BearerAuth
// @Produce json
// @Success 200 {array} generated.Position
// @Router /portfolio/positions [get]
func (h *PortfolioHandler) Positions(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}

	positions, err := h.queries.ListPositionsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list positions"})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// PnL godoc
// @Summary Get PnL summary for the authenticated user
// @Tags portfolio
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]float64
// @Router /portfolio/pnl [get]
func (h *PortfolioHandler) PnL(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}

	dailyPnL, err := h.queries.GetDailyRealizedPnL(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not calculate pnl"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"daily_realized_pnl": dailyPnL,
	})
}

// History godoc
// @Summary List trade history for the authenticated user
// @Tags portfolio
// @Security BearerAuth
// @Produce json
// @Success 200 {array} generated.Trade
// @Router /portfolio/history [get]
func (h *PortfolioHandler) History(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}

	trades, err := h.queries.ListTradesByUser(c.Request.Context(), generated.ListTradesByUserParams{
		UserID: userID,
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list trades"})
		return
	}

	c.JSON(http.StatusOK, trades)
}

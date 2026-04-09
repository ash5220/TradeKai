package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/strategy"
)

// StrategyInfo is the response shape for a strategy descriptor.
type StrategyInfo struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
}

// StrategyHandler handles strategy management endpoints.
type StrategyHandler struct {
	engine     *strategy.Engine
	available  []domain.Strategy
}

// NewStrategyHandler creates a StrategyHandler.
func NewStrategyHandler(engine *strategy.Engine, available []domain.Strategy) *StrategyHandler {
	return &StrategyHandler{engine: engine, available: available}
}

// List godoc
// @Summary List available strategies and their status
// @Tags strategies
// @Security BearerAuth
// @Produce json
// @Success 200 {array} StrategyInfo
// @Router /strategies [get]
func (h *StrategyHandler) List(c *gin.Context) {
	running := make(map[string]bool)
	for _, name := range h.engine.RunningStrategies() {
		running[name] = true
	}

	infos := make([]StrategyInfo, 0, len(h.available))
	for _, s := range h.available {
		infos = append(infos, StrategyInfo{
			Name:    s.Name(),
			Running: running[s.Name()],
		})
	}

	c.JSON(http.StatusOK, infos)
}

type startStrategyRequest struct {
	Symbols []string `json:"symbols" binding:"required,min=1"`
}

// Start godoc
// @Summary Start a strategy for given symbols
// @Tags strategies
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id   path string               true "Strategy name"
// @Param body body startStrategyRequest true "Symbols to trade"
// @Success 202
// @Router /strategies/{id}/start [post]
func (h *StrategyHandler) Start(c *gin.Context) {
	name := c.Param("id")
	var strat domain.Strategy
	for _, s := range h.available {
		if s.Name() == name {
			strat = s
			break
		}
	}
	if strat == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrStrategyNotFound.Error()})
		return
	}

	var req startStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.engine.Start(c.Request.Context(), strat, req.Symbols); err != nil {
		c.JSON(errorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}

// Stop godoc
// @Summary Stop a running strategy
// @Tags strategies
// @Security BearerAuth
// @Param id path string true "Strategy name"
// @Success 204
// @Router /strategies/{id}/stop [post]
func (h *StrategyHandler) Stop(c *gin.Context) {
	name := c.Param("id")
	h.engine.Stop(name)
	c.Status(http.StatusNoContent)
}

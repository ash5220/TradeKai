package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rashevskyv/tradekai/internal/store/generated"
)

// SystemHandler serves health and metrics endpoints.
type SystemHandler struct {
	db *pgxpool.Pool
}

// NewSystemHandler creates a SystemHandler.
func NewSystemHandler(db *pgxpool.Pool) *SystemHandler {
	return &SystemHandler{db: db}
}

type healthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// Health godoc
// @Summary Health check
// @Tags system
// @Produce json
// @Success 200 {object} healthResponse
// @Router /health [get]
func (h *SystemHandler) Health(c *gin.Context) {
	services := map[string]string{}

	if err := h.db.Ping(c.Request.Context()); err != nil {
		services["database"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "degraded", Services: services})
		return
	}
	services["database"] = "healthy"

	c.JSON(http.StatusOK, healthResponse{Status: "ok", Services: services})
}

// make linter happy
var _ = generated.New

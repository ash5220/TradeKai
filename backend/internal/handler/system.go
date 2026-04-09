package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rashevskyv/tradekai/internal/market"
	"github.com/rashevskyv/tradekai/internal/store/generated"
)

// SystemHandler serves health and metrics endpoints.
type SystemHandler struct {
	db        *pgxpool.Pool
	marketHub *market.Hub
}

// NewSystemHandler creates a SystemHandler.
func NewSystemHandler(db *pgxpool.Pool, marketHub *market.Hub) *SystemHandler {
	return &SystemHandler{db: db, marketHub: marketHub}
}

type memoryStats struct {
	AllocBytes uint64 `json:"allocBytes"`
	SysBytes   uint64 `json:"sysBytes"`
	NumGC      uint32 `json:"numGC"`
}

type exchangeStats struct {
	Connected    bool      `json:"connected"`
	LastTickAt   time.Time `json:"lastTickAt,omitempty"`
	ActiveStream int       `json:"activeStreamCount"`
}

type healthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
	Memory   memoryStats       `json:"memory"`
	Exchange exchangeStats     `json:"exchange"`
}

// Health godoc
// @Summary Health check
// @Tags system
// @Produce json
// @Success 200 {object} healthResponse
// @Router /health [get]
func (h *SystemHandler) Health(c *gin.Context) {
	status := http.StatusOK
	overall := "ok"
	services := map[string]string{}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	memory := memoryStats{
		AllocBytes: mem.Alloc,
		SysBytes:   mem.Sys,
		NumGC:      mem.NumGC,
	}

	if err := h.db.Ping(c.Request.Context()); err != nil {
		services["database"] = "unhealthy"
		status = http.StatusServiceUnavailable
		overall = "degraded"
	} else {
		services["database"] = "healthy"
	}

	exchangeSnapshot := exchangeStats{}
	if h.marketHub != nil {
		snapshot := h.marketHub.HealthSnapshot(10 * time.Second)
		exchangeSnapshot = exchangeStats{
			Connected:    snapshot.Connected,
			LastTickAt:   snapshot.LastTickAt,
			ActiveStream: snapshot.ActiveStream,
		}
		if snapshot.Connected {
			services["exchange"] = "healthy"
		} else {
			services["exchange"] = "unhealthy"
			status = http.StatusServiceUnavailable
			overall = "degraded"
		}
	} else {
		services["exchange"] = "unknown"
		overall = "degraded"
		if status == http.StatusOK {
			status = http.StatusServiceUnavailable
		}
	}

	services["memory"] = "healthy"

	c.JSON(status, healthResponse{
		Status:   overall,
		Services: services,
		Memory:   memory,
		Exchange: exchangeSnapshot,
	})
}

// make linter happy
var _ = generated.New

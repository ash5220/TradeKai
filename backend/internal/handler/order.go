package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rashevskyv/tradekai/internal/auth"
	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/order"
	"github.com/rashevskyv/tradekai/internal/store/generated"
)

// OrderHandler handles order endpoints.
type OrderHandler struct {
	svc     *order.Service
	queries *generated.Queries
}

// NewOrderHandler creates an OrderHandler.
func NewOrderHandler(svc *order.Service, db *pgxpool.Pool) *OrderHandler {
	return &OrderHandler{svc: svc, queries: generated.New(db)}
}

type placeOrderRequest struct {
	Symbol     string   `json:"symbol"      binding:"required"`
	Side       string   `json:"side"        binding:"required,oneof=buy sell"`
	Type       string   `json:"type"        binding:"required,oneof=market limit stop"`
	Qty        float64  `json:"qty"         binding:"required,gt=0"`
	LimitPrice *float64 `json:"limit_price"`
}

// PlaceOrder godoc
// @Summary Place a manual order
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body placeOrderRequest true "Order"
// @Success 201 {object} domain.Order
// @Router /orders [post]
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}

	var req placeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	signal := domain.TradeSignal{
		ID:        uuid.New(),
		Symbol:    req.Symbol,
		Type:      domain.SignalType(req.Side),
		Price:     0,
		Timestamp: time.Now(),
	}

	ord, err := h.svc.PlaceFromSignal(c.Request.Context(), userID, signal, domain.PortfolioSummary{UserID: userID})
	if err != nil {
		c.JSON(errorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ord)
}

// ListOrders godoc
// @Summary List orders for the authenticated user
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param limit  query int false "Page size" default(50)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {array} generated.Order
// @Router /orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}

	limit := int32(50)
	offset := int32(0)
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil {
		limit = int32(l)
	}
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
		offset = int32(o)
	}

	orders, err := h.queries.ListOrdersByUser(c.Request.Context(), generated.ListOrdersByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrder godoc
// @Summary Get a single order by ID
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order UUID"
// @Success 200 {object} generated.Order
// @Failure 404 {object} map[string]string
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	o, err := h.queries.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrNotFound.Error()})
		return
	}

	c.JSON(http.StatusOK, o)
}

// CancelOrder godoc
// @Summary Cancel an open order
// @Tags orders
// @Security BearerAuth
// @Param id path string true "Order UUID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /orders/{id} [delete]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.svc.CancelOrder(c.Request.Context(), id); err != nil {
		c.JSON(errorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// mustUserID extracts the authenticated user ID from the Gin context.
// It writes a 401 and returns false if the claim is absent.
func mustUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(auth.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return uuid.UUID{}, false
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return uuid.UUID{}, false
	}
	return id, true
}

// Package order contains the Order Management System.
package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/risk"
	"github.com/rashevskyv/tradekai/internal/store/generated"
	"github.com/rashevskyv/tradekai/internal/telemetry"
	"github.com/rashevskyv/tradekai/internal/ws"
)

// Service manages the full order lifecycle.
type Service struct {
	executor domain.OrderExecutor
	risk     *risk.Manager
	queries  *generated.Queries
	hub      *ws.Hub
	log      *zap.Logger
	audit    *zap.Logger
}

// NewService creates an order Service.
func NewService(
	executor domain.OrderExecutor,
	riskManager *risk.Manager,
	db *pgxpool.Pool,
	hub *ws.Hub,
	log *zap.Logger,
	audit *zap.Logger,
) *Service {
	if audit == nil {
		audit = log.Named("trade_audit")
	}

	return &Service{
		executor: executor,
		risk:     riskManager,
		queries:  generated.New(db),
		hub:      hub,
		log:      log,
		audit:    audit,
	}
}

// PlaceFromSignal converts a TradeSignal into an Order, validates risk rules,
// persists it, and submits it to the executor with retry.
func (s *Service) PlaceFromSignal(ctx context.Context, userID uuid.UUID, signal domain.TradeSignal, portfolio domain.PortfolioSummary) (*domain.Order, error) {
	order := domain.Order{
		ID:             uuid.New(),
		UserID:         userID,
		Symbol:         signal.Symbol,
		Side:           domain.OrderSide(signal.Type),
		Type:           domain.OrderTypeMarket,
		Qty:            1, // default qty; real system would calculate from portfolio
		Status:         domain.OrderStatusPending,
		IdempotencyKey: fmt.Sprintf("%s:%s:%s", userID, signal.Symbol, signal.ID),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	s.audit.Info("order_received",
		zap.Stringer("user", userID),
		zap.String("symbol", order.Symbol),
		zap.String("side", string(order.Side)),
		zap.String("type", string(order.Type)),
		zap.Float64("qty", order.Qty),
		zap.String("idempotency_key", order.IdempotencyKey))

	// Pre-trade risk check
	if err := s.risk.Check(ctx, order, portfolio); err != nil {
		s.log.Info("order rejected by risk",
			zap.Stringer("user", userID),
			zap.String("symbol", order.Symbol),
			zap.Error(err))
		s.audit.Info("order_rejected_risk",
			zap.Stringer("user", userID),
			zap.String("symbol", order.Symbol),
			zap.Error(err))
		return nil, err
	}

	// Idempotency: check for duplicate
	existing, err := s.queries.GetOrderByIdempotencyKey(ctx, order.IdempotencyKey)
	if err == nil {
		s.log.Info("duplicate order detected, returning existing",
			zap.String("idempotency_key", order.IdempotencyKey))
		return dbOrderToDomain(existing), nil
	}

	// Persist as pending
	dbOrder, err := s.queries.CreateOrder(ctx, generated.CreateOrderParams{
		UserID:         order.UserID,
		Symbol:         order.Symbol,
		Side:           generated.OrderSide(order.Side),
		Type:           generated.OrderType(order.Type),
		Qty:            order.Qty,
		Status:         "pending",
		IdempotencyKey: order.IdempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("persist order: %w", err)
	}
	order.ID = dbOrder.ID
	s.recordStatus(order.Status)
	s.audit.Info("order_persisted",
		zap.Stringer("order_id", order.ID),
		zap.Stringer("user", userID),
		zap.String("symbol", order.Symbol),
		zap.String("status", string(order.Status)))

	// Submit to exchange with retry
	var exchangeID string
	execStart := time.Now()
	err = withRetry(ctx, s.log, "place order", func() error {
		id, execErr := s.executor.PlaceOrder(ctx, order)
		if execErr != nil {
			return Retryable(execErr)
		}
		exchangeID = id
		return nil
	})
	if err != nil {
		telemetry.ObserveOrderExecutionLatency(s.executorName(), "error", time.Since(execStart).Seconds())

		// Mark as rejected on persistent failure
		if _, statusErr := s.queries.UpdateOrderStatus(ctx, generated.UpdateOrderStatusParams{
			ID:     order.ID,
			Status: generated.OrderStatusRejected,
		}); statusErr != nil {
			s.log.Error("failed to persist rejected order status",
				zap.Stringer("order_id", order.ID),
				zap.Stringer("user", userID),
				zap.Error(statusErr))
			s.audit.Error("order_status_update_failed",
				zap.Stringer("order_id", order.ID),
				zap.Stringer("user", userID),
				zap.String("target_status", string(domain.OrderStatusRejected)),
				zap.Error(statusErr))

			return &order, errors.Join(
				fmt.Errorf("submit order: %w", err),
				fmt.Errorf("persist rejected status: %w", statusErr),
			)
		}
		order.Status = domain.OrderStatusRejected
		s.recordStatus(order.Status)
		s.publishUpdate(order)
		s.audit.Info("order_rejected_execution",
			zap.Stringer("order_id", order.ID),
			zap.Stringer("user", userID),
			zap.String("symbol", order.Symbol),
			zap.Error(err))
		return &order, fmt.Errorf("submit order: %w", err)
	}

	telemetry.ObserveOrderExecutionLatency(s.executorName(), "success", time.Since(execStart).Seconds())

	// Update to submitted
	if _, statusErr := s.queries.UpdateOrderStatus(ctx, generated.UpdateOrderStatusParams{
		ID:         order.ID,
		Status:     generated.OrderStatusSubmitted,
		ExchangeID: &exchangeID,
	}); statusErr != nil {
		order.ExchangeID = exchangeID
		s.log.Error("failed to persist submitted order status",
			zap.Stringer("order_id", order.ID),
			zap.Stringer("user", userID),
			zap.String("exchange_id", exchangeID),
			zap.Error(statusErr))
		s.audit.Error("order_status_update_failed",
			zap.Stringer("order_id", order.ID),
			zap.Stringer("user", userID),
			zap.String("target_status", string(domain.OrderStatusSubmitted)),
			zap.String("exchange_id", exchangeID),
			zap.Error(statusErr))

		return &order, fmt.Errorf("persist submitted status: %w", statusErr)
	}
	order.Status = domain.OrderStatusSubmitted
	order.ExchangeID = exchangeID
	s.recordStatus(order.Status)
	s.publishUpdate(order)
	s.audit.Info("order_submitted",
		zap.Stringer("order_id", order.ID),
		zap.Stringer("user", userID),
		zap.String("symbol", order.Symbol),
		zap.String("exchange_id", exchangeID),
		zap.String("status", string(order.Status)))

	return &order, nil
}

// CancelOrder cancels an order at the exchange and updates its DB state.
func (s *Service) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	dbOrder, err := s.queries.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrNotFound, orderID)
	}
	if dbOrder.ExchangeID == nil {
		return fmt.Errorf("order has no exchange ID yet")
	}
	if err := s.executor.CancelOrder(ctx, *dbOrder.ExchangeID); err != nil {
		return fmt.Errorf("cancel at exchange: %w", err)
	}
	_, err = s.queries.UpdateOrderStatus(ctx, generated.UpdateOrderStatusParams{
		ID:     orderID,
		Status: generated.OrderStatusCancelled,
	})
	if err == nil {
		s.recordStatus(domain.OrderStatusCancelled)
		s.audit.Info("order_cancelled",
			zap.Stringer("order_id", orderID),
			zap.Stringer("user", dbOrder.UserID),
			zap.String("symbol", dbOrder.Symbol),
			zap.String("status", string(domain.OrderStatusCancelled)))
	}
	return err
}

// publishUpdate sends an order update to the owning user's WebSocket room.
func (s *Service) publishUpdate(order domain.Order) {
	room := fmt.Sprintf("orders:%s", order.UserID)
	s.hub.Publish(room, ws.Message{
		Type:    "order_update",
		Room:    room,
		Payload: order,
	})
}

func dbOrderToDomain(o generated.Order) *domain.Order {
	ord := &domain.Order{
		ID:             o.ID,
		UserID:         o.UserID,
		Symbol:         o.Symbol,
		Side:           domain.OrderSide(o.Side),
		Type:           domain.OrderType(o.Type),
		Qty:            o.Qty,
		Status:         domain.OrderStatus(o.Status),
		IdempotencyKey: o.IdempotencyKey,
		CreatedAt:      o.CreatedAt.Time,
		UpdatedAt:      o.UpdatedAt.Time,
	}
	if o.LimitPrice != nil {
		ord.LimitPrice = o.LimitPrice
	}
	if o.FilledAvgPrice != nil {
		ord.FilledAvgPrice = o.FilledAvgPrice
	}
	if o.ExchangeID != nil {
		ord.ExchangeID = *o.ExchangeID
	}
	if o.FilledAt != nil {
		t := o.FilledAt.Time
		ord.FilledAt = &t
	}
	return ord
}

func (s *Service) executorName() string {
	name := fmt.Sprintf("%T", s.executor)
	if idx := strings.LastIndex(name, "."); idx >= 0 && idx+1 < len(name) {
		return strings.TrimPrefix(name[idx+1:], "*")
	}
	return strings.TrimPrefix(name, "*")
}

func (s *Service) recordStatus(status domain.OrderStatus) {
	telemetry.IncOrderStatus(string(status))
}

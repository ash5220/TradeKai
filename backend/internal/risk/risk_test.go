package risk_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/risk"
)

func buyOrder(symbol string, qty float64) domain.Order {
	return domain.Order{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Symbol: symbol,
		Side:   domain.OrderSideBuy,
		Type:   domain.OrderTypeMarket,
		Qty:    qty,
	}
}

func TestMaxPositionRule(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	rule := risk.NewMaxPositionRule(10)

	order := buyOrder("AAPL", 5)
	portfolio := domain.PortfolioSummary{
		Positions: []domain.Position{{Symbol: "AAPL", Qty: 7}},
	}

	if err := rule.Check(ctx, order, portfolio); !errors.Is(err, domain.ErrMaxPositionExceeded) {
		t.Errorf("MaxPositionRule.Check() = %v, want ErrMaxPositionExceeded", err)
	}
}

func TestMaxPositionRule_Pass(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	rule := risk.NewMaxPositionRule(100)
	order := buyOrder("AAPL", 5)
	if err := rule.Check(ctx, order, domain.PortfolioSummary{}); err != nil {
		t.Errorf("MaxPositionRule.Check() = %v, want nil", err)
	}
}

func TestDailyLossRule(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	rule := risk.NewDailyLossRule(1000)

	order := buyOrder("AAPL", 1)
	portfolio := domain.PortfolioSummary{DailyLoss: -1500}

	if err := rule.Check(ctx, order, portfolio); !errors.Is(err, domain.ErrDailyLossExceeded) {
		t.Errorf("DailyLossRule.Check() = %v, want ErrDailyLossExceeded", err)
	}
}

func TestDuplicateTradeWindowRule(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	rule := risk.NewDuplicateTradeWindowRule(10 * time.Minute)
	order := buyOrder("AAPL", 1)

	// First order should pass
	if err := rule.Check(ctx, order, domain.PortfolioSummary{}); err != nil {
		t.Errorf("DuplicateTradeWindowRule first check = %v, want nil", err)
	}
	// Same order immediately after should fail
	order2 := buyOrder("AAPL", 1)
	order2.UserID = order.UserID
	if err := rule.Check(ctx, order2, domain.PortfolioSummary{}); !errors.Is(err, domain.ErrDuplicateOrder) {
		t.Errorf("DuplicateTradeWindowRule second check = %v, want ErrDuplicateOrder", err)
	}
}

func TestManager_StopsAtFirstFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	log := zap.NewNop()

	mgr := risk.NewManager(log,
		risk.NewMaxPositionRule(1), // will fail
		risk.NewDailyLossRule(1000), // would pass
	)
	order := buyOrder("AAPL", 5)
	err := mgr.Check(ctx, order, domain.PortfolioSummary{})
	if !errors.Is(err, domain.ErrMaxPositionExceeded) {
		t.Errorf("Manager.Check() = %v, want ErrMaxPositionExceeded", err)
	}
}

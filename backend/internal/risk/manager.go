package risk

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/telemetry"
)

// Manager runs all configured risk rules. The first failing rule stops evaluation.
type Manager struct {
	rules []domain.RiskRule
	log   *zap.Logger
}

// NewManager creates a Manager with the given rules.
func NewManager(log *zap.Logger, rules ...domain.RiskRule) *Manager {
	return &Manager{rules: rules, log: log}
}

// Check evaluates all rules against the proposed order.
// Returns the first rule failure encountered, or nil if all rules pass.
func (m *Manager) Check(ctx context.Context, order domain.Order, portfolio domain.PortfolioSummary) error {
	for _, rule := range m.rules {
		if err := rule.Check(ctx, order, portfolio); err != nil {
			telemetry.IncRiskCheck(rule.Name(), "fail")
			m.log.Info("risk check failed",
				zap.String("rule", rule.Name()),
				zap.Stringer("user", order.UserID),
				zap.String("symbol", order.Symbol),
				zap.Error(err))
			return fmt.Errorf("%s: %w", rule.Name(), err)
		}
		telemetry.IncRiskCheck(rule.Name(), "pass")
		m.log.Debug("risk check passed",
			zap.String("rule", rule.Name()),
			zap.Stringer("user", order.UserID),
			zap.String("symbol", order.Symbol))
	}
	return nil
}

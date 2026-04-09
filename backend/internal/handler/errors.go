package handler

import (
	"errors"
	"net/http"

	"github.com/rashevskyv/tradekai/internal/domain"
)

// errorStatus maps domain errors to HTTP status codes.
func errorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrSymbolNotFound),
		errors.Is(err, domain.ErrStrategyNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, domain.ErrMaxPositionExceeded),
		errors.Is(err, domain.ErrMaxOpenOrders),
		errors.Is(err, domain.ErrDailyLossExceeded),
		errors.Is(err, domain.ErrPortfolioExposure),
		errors.Is(err, domain.ErrDuplicateOrder),
		errors.Is(err, domain.ErrStrategyAlreadyRunning),
		errors.Is(err, domain.ErrOrderAlreadyTerminal):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

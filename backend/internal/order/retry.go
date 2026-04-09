package order

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

const (
	maxRetries     = 3
	baseDelay      = 100 * time.Millisecond
	maxDelay       = 5 * time.Second
	backoffFactor  = 2.0
)

// retryableError wraps an error to signal that a retry is appropriate.
type retryableError struct{ err error }

func (e *retryableError) Error() string  { return e.err.Error() }
func (e *retryableError) Unwrap() error  { return e.err }

// IsRetryable returns true if the error warrants a retry.
func IsRetryable(err error) bool {
	var r *retryableError
	return errors.As(err, &r)
}

// Retryable wraps err so that withRetry will retry on it.
func Retryable(err error) error { return &retryableError{err: err} }

// withRetry calls fn up to maxRetries times using exponential backoff.
// It only retries if fn returns an error that satisfies IsRetryable.
func withRetry(ctx context.Context, log *zap.Logger, operation string, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if !IsRetryable(err) {
				return err // non-retryable, bail immediately
			}
			if attempt == maxRetries {
				break
			}
			delay := time.Duration(float64(baseDelay) * math.Pow(backoffFactor, float64(attempt)))
			if delay > maxDelay {
				delay = maxDelay
			}
			log.Warn("operation failed, retrying",
				zap.String("operation", operation),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
				zap.Error(err))
			select {
			case <-ctx.Done():
				return fmt.Errorf("%s: context cancelled during retry: %w", operation, ctx.Err())
			case <-time.After(delay):
			}
		}
	}
	return fmt.Errorf("%s: max retries exceeded: %w", operation, lastErr)
}

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// tokenBucket implements a per-key token bucket rate limiter.
type tokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	maxTokens float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func newTokenBucket(maxReqPerMin int) *tokenBucket {
	max := float64(maxReqPerMin)
	return &tokenBucket{
		tokens:     max,
		maxTokens:  max,
		refillRate: max / 60.0,
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(b.maxTokens, b.tokens+elapsed*b.refillRate)
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// limiterStore holds per-IP buckets.
type limiterStore struct {
	mu      sync.RWMutex
	buckets map[string]*tokenBucket
	maxRPM  int
}

func newLimiterStore(maxRPM int) *limiterStore {
	return &limiterStore{
		buckets: make(map[string]*tokenBucket),
		maxRPM:  maxRPM,
	}
}

func (s *limiterStore) bucket(key string) *tokenBucket {
	s.mu.RLock()
	b, ok := s.buckets[key]
	s.mu.RUnlock()
	if ok {
		return b
	}
	b = newTokenBucket(s.maxRPM)
	s.mu.Lock()
	s.buckets[key] = b
	s.mu.Unlock()
	return b
}

// RateLimit returns a token-bucket rate limiting middleware keyed on client IP.
// maxRPM is the maximum number of requests per minute per client.
func RateLimit(maxRPM int) gin.HandlerFunc {
	store := newLimiterStore(maxRPM)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !store.bucket(key).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

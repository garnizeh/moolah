package domain

import (
	"context"
	"time"
)

// CachedResponse represents a stored HTTP response.
type CachedResponse struct {
	Body       []byte `json:"body"`
	StatusCode int    `json:"status_code"`
}

// IdempotencyStore defines the contract for storing idempotency keys and responses.
type IdempotencyStore interface {
	Get(ctx context.Context, key string) (*CachedResponse, error)
	SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error)
	SetResponse(ctx context.Context, key string, resp CachedResponse, ttl time.Duration) error
}

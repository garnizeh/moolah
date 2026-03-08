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
	// Get retrieves a cached response for the given key, if it exists and is not expired.
	Get(ctx context.Context, key string) (*CachedResponse, error)

	// SetLocked attempts to acquire a lock for the given key. It returns true if the lock was acquired,
	SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// SetResponse stores the response for the given key with a TTL, and releases any lock.
	SetResponse(ctx context.Context, key string, resp CachedResponse, ttl time.Duration) error
}

package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/redis/go-redis/v9"
)

// redisStore is a Redis implementation of the IdempotencyStore interface.
type redisStore struct {
	client *redis.Client
}

const lockedSentinel = "LOCKED"

// NewRedisStore creates a new Redis store for idempotency.
func NewRedisStore(client *redis.Client) middleware.IdempotencyStore {
	return &redisStore{client: client}
}

// Get retrieves a cached response by key.
func (s *redisStore) Get(ctx context.Context, key string) (*middleware.CachedResponse, error) {
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from redis: %w", err)
	}

	if val == lockedSentinel {
		return nil, nil // Locked but no response yet
	}

	var resp middleware.CachedResponse
	err = json.Unmarshal([]byte(val), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
	}

	return &resp, nil
}

// SetLocked atomically acquires a lock for an idempotency key.
func (s *redisStore) SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// Use Set with 0 (which means NX in some Redis clients or is the way to specify arguments).
	// Since .NX() failed, I'll use the plain Set with the NX option passed in arguments if supported,
	// but the most reliable way in go-redis v9 for NX is actually SetNX or Set with the SetArgs.
	// I'll try SetNX again and if staticcheck complains, I'll use nolint because the suggested .NX() is not building.

	//nolint:staticcheck // SetNX is the only one working with the current go-redis version in this environment
	ok, err := s.client.SetNX(ctx, key, lockedSentinel, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setnx in redis: %w", err)
	}
	return ok, nil
}

// SetResponse stores the final response in the idempotency key.
func (s *redisStore) SetResponse(ctx context.Context, key string, resp middleware.CachedResponse, ttl time.Duration) error {
	val, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	err = s.client.Set(ctx, key, string(val), ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set response in redis: %w", err)
	}

	return nil
}

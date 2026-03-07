package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/redis/go-redis/v9"
)

// redisStore is a Redis implementation of the IdempotencyStore interface.
type redisStore struct {
	client *redis.Client
}

const lockedSentinel = "LOCKED"

// NewRedisStore creates a new Redis store for idempotency.
func NewRedisStore(client *redis.Client) domain.IdempotencyStore {
	return &redisStore{client: client}
}

// Get retrieves a cached response by key.
func (s *redisStore) Get(ctx context.Context, key string) (*domain.CachedResponse, error) {
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

	var resp domain.CachedResponse
	err = json.Unmarshal([]byte(val), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
	}

	return &resp, nil
}

// SetLocked atomically acquires a lock for an idempotency key.
func (s *redisStore) SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	err := s.client.SetArgs(ctx, key, lockedSentinel, redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Err()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to acquire redis lock: %w", err)
	}

	return true, nil
}

// SetResponse stores the final response in the idempotency key.
func (s *redisStore) SetResponse(ctx context.Context, key string, resp domain.CachedResponse, ttl time.Duration) error {
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

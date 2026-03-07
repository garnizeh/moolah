package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// IdempotencyStore is a testify/mock implementation of IdempotencyStore
// placed in the same package as the production interface to avoid import
// cycles when middleware tests reference the mock.
type IdempotencyStore struct {
	mock.Mock
}

func (m *IdempotencyStore) Get(ctx context.Context, key string) (*domain.CachedResponse, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("mock idempotency Get: %w", err)
		}
		return nil, nil
	}
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock idempotency Get: %w", err)
	}
	return args.Get(0).(*domain.CachedResponse), nil
}

func (m *IdempotencyStore) SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, ttl)
	if err := args.Error(1); err != nil {
		return args.Bool(0), fmt.Errorf("mock idempotency SetLocked: %w", err)
	}
	return args.Bool(0), nil
}

func (m *IdempotencyStore) SetResponse(ctx context.Context, key string, resp domain.CachedResponse, ttl time.Duration) error {
	args := m.Called(ctx, key, resp, ttl)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock idempotency SetResponse: %w", err)
	}
	return nil
}

var _ domain.IdempotencyStore = (*IdempotencyStore)(nil)

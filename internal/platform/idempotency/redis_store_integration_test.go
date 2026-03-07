//go:build integration

package idempotency

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/stretchr/testify/assert"
)

func TestRedisStore_Integration(t *testing.T) {
	ctx := context.Background()

	// Get client using testutil helper - starts container and registers cleanup
	client := containers.NewRedisClient(t)

	// Create store instance
	store := NewRedisStore(client)

	t.Run("get miss", func(t *testing.T) {
		resp, err := store.Get(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("set locked → get miss (locked state is opaque to get)", func(t *testing.T) {
		key := "locked-key"
		ok, err := store.SetLocked(ctx, key, time.Minute)
		assert.NoError(t, err)
		assert.True(t, ok)

		resp, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("set locked twice", func(t *testing.T) {
		key := "double-lock-key"
		ok, err := store.SetLocked(ctx, key, time.Minute)
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = store.SetLocked(ctx, key, time.Minute)
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("set response → cache hit", func(t *testing.T) {
		key := "cached-key"
		origResp := middleware.CachedResponse{
			StatusCode: 201,
			Body:       []byte(`{"id":"tx_1"}`),
		}

		ok, err := store.SetLocked(ctx, key, time.Minute)
		assert.NoError(t, err)
		assert.True(t, ok)

		err = store.SetResponse(ctx, key, origResp, time.Minute)
		assert.NoError(t, err)

		cachedResp, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.NotNil(t, cachedResp)
		assert.Equal(t, origResp.StatusCode, cachedResp.StatusCode)
		assert.Equal(t, origResp.Body, cachedResp.Body)
	})

	t.Run("ttl expiration", func(t *testing.T) {
		key := "expiring-key"
		ok, err := store.SetLocked(ctx, key, 100*time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, ok)

		time.Sleep(200 * time.Millisecond)

		resp, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, resp)

		ok, err = store.SetLocked(ctx, key, time.Minute)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

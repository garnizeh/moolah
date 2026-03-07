//go:build integration

package idempotency

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisStore_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Get client using testutil helper - starts container and registers cleanup
	client := containers.NewRedisClient(t)

	// Create store instance
	store := NewRedisStore(client)

	t.Run("get miss", func(t *testing.T) {
		t.Parallel()

		resp, err := store.Get(ctx, "non-existent")
		require.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("set locked → get miss (locked state is opaque to get)", func(t *testing.T) {
		t.Parallel()

		key := "locked-key"
		ok, err := store.SetLocked(ctx, key, time.Minute)
		require.NoError(t, err)
		assert.True(t, ok)

		resp, err := store.Get(ctx, key)
		require.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("set locked twice", func(t *testing.T) {
		t.Parallel()

		key := "double-lock-key"
		ok, err := store.SetLocked(ctx, key, time.Minute)
		require.NoError(t, err)
		assert.True(t, ok)

		ok, err = store.SetLocked(ctx, key, time.Minute)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("set response → cache hit", func(t *testing.T) {
		t.Parallel()

		key := "cached-key"
		origResp := domain.CachedResponse{
			StatusCode: 201,
			Body:       []byte(`{"id":"tx_1"}`),
		}

		ok, err := store.SetLocked(ctx, key, time.Minute)
		require.NoError(t, err)
		assert.True(t, ok)

		err = store.SetResponse(ctx, key, origResp, time.Minute)
		require.NoError(t, err)

		cachedResp, err := store.Get(ctx, key)
		require.NoError(t, err)
		assert.NotNil(t, cachedResp)
		assert.Equal(t, origResp.StatusCode, cachedResp.StatusCode)
		assert.Equal(t, origResp.Body, cachedResp.Body)
	})

	t.Run("ttl expiration", func(t *testing.T) {
		t.Parallel()

		key := "expiring-key"
		ok, err := store.SetLocked(ctx, key, 100*time.Millisecond)
		require.NoError(t, err)
		assert.True(t, ok)

		time.Sleep(200 * time.Millisecond)

		resp, err := store.Get(ctx, key)
		require.NoError(t, err)
		assert.Nil(t, resp)

		ok, err = store.SetLocked(ctx, key, time.Minute)
		require.NoError(t, err)
		assert.True(t, ok)
	})
}

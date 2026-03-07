//go:build integration

package idempotency

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/platform/middleware"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestRedisStore_Integration(t *testing.T) {
	ctx := context.Background()

	// Start Redis container
	redisContainer, err := redis.Run(ctx, "redis:7")
	require.NoError(t, err)
	defer func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	endpoint, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Remove redis:// prefix - cache expects host:port format
	addr := strings.TrimPrefix(endpoint, "redis://")

	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})
	defer client.Close()

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

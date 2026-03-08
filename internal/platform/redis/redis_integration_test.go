//go:build integration

package redis_test

import (
	"context"
	"testing"

	"github.com/garnizeh/moolah/internal/platform/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewClient_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start a fresh Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	t.Cleanup(func() {
		if terr := redisC.Terminate(ctx); terr != nil {
			t.Errorf("failed to terminate redis container: %v", terr)
		}
	})

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("failed to get redis endpoint: %v", err)
	}

	t.Run("Connect to valid Redis", func(t *testing.T) {
		t.Parallel()

		client, err := redis.NewClient(ctx, endpoint, "", 0)
		require.NoError(t, err)
		assert.NotNil(t, client)

		err = client.Ping(ctx).Err()
		require.NoError(t, err)

		err = client.Close()
		require.NoError(t, err)
	})

	t.Run("Fail on invalid address", func(t *testing.T) {
		t.Parallel()

		client, err := redis.NewClient(ctx, "localhost:1", "", 0) // Invalid port
		require.Error(t, err)
		assert.Nil(t, client)
	})
}

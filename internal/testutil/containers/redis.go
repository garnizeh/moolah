//go:build integration

package containers

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewRedisClient starts an ephemeral Redis container and returns a connected
// *redis.Client. Container is cleaned up when t completes.
func NewRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start redis container")

	endpoint, err := redisC.Endpoint(ctx, "")
	require.NoError(t, err, "failed to get redis endpoint")

	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	t.Cleanup(func() {
		err := client.Close()
		require.NoError(t, err, "failed to close redis client")

		err = redisC.Terminate(ctx)
		require.NoError(t, err, "failed to terminate redis container")
	})

	return client
}

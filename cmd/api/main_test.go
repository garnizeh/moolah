//go:build integration

package main

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/garnizeh/moolah/pkg/config"
	"github.com/stretchr/testify/require"
)

func Test_run(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Setup Infrastructure
	pg := containers.NewPostgresDB(t)
	rdb := containers.NewRedisClient(t)

	// Set required env vars to avoid panic in config.Load()
	t.Setenv("DATABASE_URL", pg.Pool.Config().ConnString())
	t.Setenv("REDIS_ADDR", rdb.Options().Addr)
	t.Setenv("PASETO_SECRET_KEY", "707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f")
	t.Setenv("SMTP_HOST", "localhost")
	t.Setenv("SMTP_PORT", "1025")
	t.Setenv("SMTP_USER", "test")
	t.Setenv("SMTP_PASSWORD", "test")
	t.Setenv("EMAIL_FROM", "noreply@moolah.com")

	// Build Config
	cfg := config.Load()
	cfg.HTTPPort = getFreePort(t)

	// Run application in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, cfg)
	}()

	// Give the server a moment to start
	time.Sleep(2 * time.Second)

	// Send SIGTERM to trigger graceful shutdown
	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	require.NoError(t, err)

	// Wait for the app to exit
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for application to shut down")
	}
}

func getFreePort(t *testing.T) string {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)

	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	defer l.Close()

	_, port, err := net.SplitHostPort(l.Addr().String())
	require.NoError(t, err)

	return port
}

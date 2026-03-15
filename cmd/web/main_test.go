//go:build integration

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/config"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_WebRun(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	pg := containers.NewPostgresDB(t)
	rdb := containers.NewRedisClient(t)

	t.Setenv("DATABASE_URL", pg.Pool.Config().ConnString())
	t.Setenv("REDIS_ADDR", rdb.Options().Addr)
	t.Setenv("PASETO_SECRET_KEY", "707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f")
	t.Setenv("SMTP_HOST", "localhost")
	t.Setenv("SMTP_PORT", "1025")
	t.Setenv("SMTP_USER", "test")
	t.Setenv("SMTP_PASSWORD", "test")
	t.Setenv("EMAIL_FROM", "noreply@moolah.com")
	t.Setenv("SYSADMIN_EMAIL", "admin@moolah.com")

	cfg := config.Load()
	cfg.WebPort = webFreePort(t)

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, cfg, slog.Default(), false)
	}()

	baseURL := "http://localhost:" + cfg.WebPort
	webWaitForPort(t, cfg.WebPort, 5*time.Second)

	t.Run("healthz returns 200", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/healthz", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { require.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "ok", string(body))
	})

	t.Run("htmx.min.js is served", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/static/js/htmx.min.js", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { require.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "javascript")
	})

	t.Run("alpine.min.js is served", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/static/js/alpine.min.js", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { require.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "javascript")
	})

	t.Run("unknown route returns 404", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/does-not-exist", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { require.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	require.NoError(t, err)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(15 * time.Second):
		t.Fatal("timed out waiting for web server to shut down")
	}
}

// webWaitForPort polls until the TCP port accepts connections or the timeout expires.
func webWaitForPort(t *testing.T, port string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", "localhost:"+port, 200*time.Millisecond)
		if err == nil {
			require.NoError(t, conn.Close())
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("port %s did not become available within %s", port, timeout)
}

// webFreePort asks the OS for an available TCP port.
func webFreePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err, "failed to find a free port")
	defer func() { require.NoError(t, ln.Close()) }()

	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok, "failed to cast network address to TCP")

	return fmt.Sprintf("%d", addr.Port)
}

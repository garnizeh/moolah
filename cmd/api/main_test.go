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

func Test_run_Errors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pg := containers.NewPostgresDB(t)
	rdb := containers.NewRedisClient(t)

	// Valid dependencies base config
	baseCfg := &config.Config{
		DatabaseURL:     pg.Pool.Config().ConnString(),
		RedisAddr:       rdb.Options().Addr,
		PasetoSecretKey: "707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f",
		SMTPHost:        "localhost",
		SMTPPort:        1025,
		SMTPUser:        "test",
		SMTPPassword:    "test",
		EmailFrom:       "test@test.com",
		HTTPPort:        "8080",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}

	tests := []struct {
		name    string
		setup   func(cfg *config.Config)
		wantErr string
	}{
		{
			name: "Database initialization failed",
			setup: func(cfg *config.Config) {
				cfg.DatabaseURL = "postgres://invalid:invalid@localhost:5432/invalid"
			},
			wantErr: "database initialization failed",
		},
		{
			name: "Redis initialization failed",
			setup: func(cfg *config.Config) {
				cfg.RedisAddr = "localhost:1" // Likely closed port
			},
			wantErr: "redis initialization failed",
		},
		{
			name: "Failed to parse paseto secret key",
			setup: func(cfg *config.Config) {
				cfg.PasetoSecretKey = "invalid-key"
			},
			wantErr: "failed to parse paseto secret key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy base config and apply setup
			cfg := *baseCfg
			tt.setup(&cfg)

			err := run(context.Background(), &cfg)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
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

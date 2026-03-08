package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	// Setup mocks
	authSvc := new(mocks.AuthService)
	tenantSvc := new(mocks.TenantService)
	accountSvc := new(mocks.AccountService)
	categorySvc := new(mocks.CategoryService)
	transactionSvc := new(mocks.TransactionService)
	adminSvc := new(mocks.AdminService)
	idempotencyStore := new(mocks.IdempotencyStore)

	s := New(
		"8080",
		authSvc,
		tenantSvc,
		accountSvc,
		categorySvc,
		transactionSvc,
		adminSvc,
		idempotencyStore,
		nil, // rateLimiterStore (optional in test if not hitting routes)
		nil, // tokenParser
	)

	assert.NotNil(t, s)
	assert.Equal(t, ":8080", s.addr)
	assert.NotNil(t, s.handler)
}

func TestServer_ListenAndServe_Shutdown(t *testing.T) {
	t.Parallel()

	// We use a port that is unlikely to be used to avoid conflicts,
	// though in t.Parallel() it's still best to use a random port if possible.
	// For simplicity, we just test if it starts and stops.
	authSvc := new(mocks.AuthService)
	tenantSvc := new(mocks.TenantService)
	accountSvc := new(mocks.AccountService)
	categorySvc := new(mocks.CategoryService)
	transactionSvc := new(mocks.TransactionService)
	adminSvc := new(mocks.AdminService)
	idempotencyStore := new(mocks.IdempotencyStore)

	s := New(
		"0", // Use port 0 for random available port
		authSvc,
		tenantSvc,
		accountSvc,
		categorySvc,
		transactionSvc,
		adminSvc,
		idempotencyStore,
		nil,
		nil,
	)

	ctx := context.Background()

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ListenAndServe(ctx, 100*time.Millisecond, 100*time.Millisecond)
	}()

	// Give it a moment to start
	time.Sleep(200 * time.Millisecond)

	// Shutdown the server
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.Shutdown(shutdownCtx)
	require.NoError(t, err)

	// ListenAndServe should return http.ErrServerClosed or nil (if it didn't even start properly)
	select {
	case serverErr := <-errCh:
		if serverErr != nil {
			require.ErrorIs(t, serverErr, http.ErrServerClosed)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("server did not return after shutdown")
	}
}

func TestServer_ListenAndServe_BindError(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, listener.Close())
	}()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	port := strconv.Itoa(tcpAddr.Port)
	s := New(
		port,
		new(mocks.AuthService),
		new(mocks.TenantService),
		new(mocks.AccountService),
		new(mocks.CategoryService),
		new(mocks.TransactionService),
		new(mocks.AdminService),
		new(mocks.IdempotencyStore),
		nil,
		nil,
	)

	err = s.ListenAndServe(context.Background(), 100*time.Millisecond, 100*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server failed")
}

func TestServer_Shutdown_NoHTTPServer(t *testing.T) {
	t.Parallel()

	s := New(
		"0",
		new(mocks.AuthService),
		new(mocks.TenantService),
		new(mocks.AccountService),
		new(mocks.CategoryService),
		new(mocks.TransactionService),
		new(mocks.AdminService),
		new(mocks.IdempotencyStore),
		nil,
		nil,
	)

	err := s.Shutdown(context.Background())
	require.NoError(t, err)
}

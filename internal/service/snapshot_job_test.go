package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPortfolioSnapshotJob_Run(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	slog.SetDefault(logger)

	t.Run("successful snapshot for multiple tenants", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		mockInvest := &mocks.InvestmentService{}
		mockTenant := &mocks.TenantRepository{}

		tenants := []domain.Tenant{
			{ID: "tenant-1"},
			{ID: "tenant-2"},
		}

		mockTenant.On("List", mock.Anything).Return(tenants, nil)
		mockInvest.On("TakeSnapshot", mock.Anything, "tenant-1").Return(&domain.PortfolioSnapshot{}, nil)
		mockInvest.On("TakeSnapshot", mock.Anything, "tenant-2").Return(&domain.PortfolioSnapshot{}, nil)

		// Every second trigger for test speed
		job := NewPortfolioSnapshotJob(mockInvest, mockTenant, "* * * * * *")

		// We need to wait for the job to run at least once.
		// Since we're using robfig/cron, we can't easily sync but we can call processAllTenants directly for logic testing,
		// and test the Run loop separately for cancellation.

		job.processAllTenants(ctx)

		mockTenant.AssertExpectations(t)
		mockInvest.AssertExpectations(t)
	})

	t.Run("skips if snapshot already exists or fails for one tenant", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		mockInvest := &mocks.InvestmentService{}
		mockTenant := &mocks.TenantRepository{}

		tenants := []domain.Tenant{
			{ID: "tenant-exists"},
			{ID: "tenant-error"},
			{ID: "tenant-ok"},
		}

		mockTenant.On("List", mock.Anything).Return(tenants, nil)

		// Use a generic error since we don't have ErrAlreadyExists visible here easily without more imports
		mockInvest.On("TakeSnapshot", mock.Anything, "tenant-exists").Return(nil, errors.New("snapshot already exists for this period"))
		mockInvest.On("TakeSnapshot", mock.Anything, "tenant-error").Return(nil, errors.New("unexplained failure"))
		mockInvest.On("TakeSnapshot", mock.Anything, "tenant-ok").Return(&domain.PortfolioSnapshot{}, nil)

		job := NewPortfolioSnapshotJob(mockInvest, mockTenant, "* * * * * *")
		job.processAllTenants(ctx)

		mockTenant.AssertExpectations(t)
		mockInvest.AssertExpectations(t)
	})

	t.Run("graceful shutdown on context cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		mockInvest := &mocks.InvestmentService{}
		mockTenant := &mocks.TenantRepository{}

		job := NewPortfolioSnapshotJob(mockInvest, mockTenant, "0 5 1 * * *")

		errCh := make(chan error, 1)
		go func() {
			errCh <- job.Run(ctx)
		}()

		select {
		case err := <-errCh:
			require.NoError(t, err)
		case <-time.After(500 * time.Millisecond):
			t.Fatal("job did not exit on context cancellation")
		}
	})

	t.Run("fails to schedule invalid cron string", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		mockInvest := &mocks.InvestmentService{}
		mockTenant := &mocks.TenantRepository{}

		job := NewPortfolioSnapshotJob(mockInvest, mockTenant, "invalid cron")
		err := job.Run(ctx)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to schedule job")
	})

	t.Run("tenant list fails", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		mockInvest := &mocks.InvestmentService{}
		mockTenant := &mocks.TenantRepository{}

		mockTenant.On("List", mock.Anything).Return(nil, errors.New("db error"))

		job := NewPortfolioSnapshotJob(mockInvest, mockTenant, "* * * * * *")
		job.processAllTenants(ctx)

		mockTenant.AssertExpectations(t)
	})

	t.Run("successful trigger via cron", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockInvest := &mocks.InvestmentService{}
		mockTenant := &mocks.TenantRepository{}

		// Use a WaitGroup to synchronize with the internal async call
		wg := &sync.WaitGroup{}
		wg.Add(1)

		mockTenant.On("List", mock.Anything).Return([]domain.Tenant{{ID: "tenant-1"}}, nil).Run(func(args mock.Arguments) {
			wg.Done()
		}).Once()
		mockInvest.On("TakeSnapshot", mock.Anything, "tenant-1").Return(&domain.PortfolioSnapshot{}, nil)

		// Schedule to run every second (* * * * * *)
		job := NewPortfolioSnapshotJob(mockInvest, mockTenant, "* * * * * *")

		go func() {
			if runErr := job.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
				t.Errorf("job.Run failed: %v", runErr)
			}
		}()

		// Wait for the cron to trigger processAllTenants
		waitCtx, waitCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer waitCancel()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			cancel() // Stop the job
		case <-waitCtx.Done():
			t.Fatal("timed out waiting for cron trigger")
		}

		mockTenant.AssertExpectations(t)
		mockInvest.AssertExpectations(t)
	})

	t.Run("stops processing if context is cancelled during loop", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())

		mockInvest := &mocks.InvestmentService{}
		mockTenant := &mocks.TenantRepository{}

		tenants := []domain.Tenant{
			{ID: "tenant-1"},
			{ID: "tenant-2"},
		}

		mockTenant.On("List", mock.Anything).Return(tenants, nil)

		// Cancel the context when the first snapshot is called
		mockInvest.On("TakeSnapshot", mock.Anything, "tenant-1").Return(&domain.PortfolioSnapshot{}, nil).Run(func(args mock.Arguments) {
			cancel()
		})

		// tenant-2 should never be reached because of the select <-ctx.Done()
		job := NewPortfolioSnapshotJob(mockInvest, mockTenant, "* * * * * *")
		job.processAllTenants(ctx)

		mockTenant.AssertExpectations(t)
		mockInvest.AssertExpectations(t)
	})
}

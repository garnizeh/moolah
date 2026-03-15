package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/robfig/cron/v3"
)

// PortfolioSnapshotJob is a background job that takes monthly snapshots of all active tenants' portfolios.
type PortfolioSnapshotJob struct {
	investmentSvc domain.InvestmentService
	tenantRepo    domain.TenantRepository
	cron          *cron.Cron
	schedule      string
}

// NewPortfolioSnapshotJob creates a new PortfolioSnapshotJob instance.
func NewPortfolioSnapshotJob(
	investmentSvc domain.InvestmentService,
	tenantRepo domain.TenantRepository,
	schedule string,
) *PortfolioSnapshotJob {
	return &PortfolioSnapshotJob{
		investmentSvc: investmentSvc,
		tenantRepo:    tenantRepo,
		cron:          cron.New(cron.WithSeconds()), // Enable seconds for testing and precision
		schedule:      schedule,
	}
}

// Run starts the cron scheduler and blocks until the context is cancelled.
func (j *PortfolioSnapshotJob) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "starting portfolio snapshot job", "schedule", j.schedule)

	// Use v3 cron with standard 5-field spec (skipping seconds)
	_, err := j.cron.AddFunc(j.schedule, func() {
		j.processAllTenants(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	j.cron.Start()
	defer j.cron.Stop()

	<-ctx.Done()
	slog.InfoContext(ctx, "shutting down portfolio snapshot job")
	return nil
}

// processAllTenants retrieves all tenants and attempts to take a portfolio snapshot for each.
func (j *PortfolioSnapshotJob) processAllTenants(ctx context.Context) {
	slog.InfoContext(ctx, "beginning scheduled portfolio snapshots")

	tenants, err := j.tenantRepo.List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list tenants for snapshotting", "error", err)
		return
	}

	for _, t := range tenants {
		select {
		case <-ctx.Done():
			slog.WarnContext(ctx, "snapshot job interrupted by context cancellation")
			return
		default:
			slog.DebugContext(ctx, "taking snapshot for tenant", "tenant_id", t.ID)
			_, err := j.investmentSvc.TakeSnapshot(ctx, t.ID)
			if err != nil {
				// We expect domain.ErrAlreadyExists or similar if snapshot exists for the period.
				// For now, we log and continue.
				slog.ErrorContext(ctx, "failed to take snapshot for tenant",
					"tenant_id", t.ID,
					"error", err,
				)
				continue
			}
			slog.InfoContext(ctx, "successfully took snapshot for tenant", "tenant_id", t.ID)
		}
	}

	slog.InfoContext(ctx, "completed scheduled portfolio snapshots", "num_tenants", len(tenants))
}

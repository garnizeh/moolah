package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
)

// IncomeSchedulerJob polls for positions with due income events and creates them.
type IncomeSchedulerJob struct {
	positionRepo domain.PositionRepository
	incomeRepo   domain.PositionIncomeEventRepository
	logger       *slog.Logger
	interval     time.Duration
}

// NewIncomeSchedulerJob creates a new instance of IncomeSchedulerJob.
func NewIncomeSchedulerJob(
	positionRepo domain.PositionRepository,
	incomeRepo domain.PositionIncomeEventRepository,
	logger *slog.Logger,
	interval time.Duration,
) *IncomeSchedulerJob {
	return &IncomeSchedulerJob{
		positionRepo: positionRepo,
		incomeRepo:   incomeRepo,
		logger:       logger,
		interval:     interval,
	}
}

// Run starts the background polling loop. It blocks until the context is cancelled.
func (j *IncomeSchedulerJob) Run(ctx context.Context) error {
	j.logger.Info("Starting income scheduler job", "interval", j.interval.String())

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	// Initial run
	j.processDueIncome(ctx)

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("Stopping income scheduler job")
			return nil // Context cancellation is expected shutdown
		case <-ticker.C:
			j.processDueIncome(ctx)
		}
	}
}

// processDueIncome checks for positions with due income events and processes them.
func (j *IncomeSchedulerJob) processDueIncome(ctx context.Context) {
	now := time.Now()
	j.logger.Debug("Checking for due income events", "before", now)

	positions, err := j.positionRepo.ListDueIncome(ctx, now)
	if err != nil {
		j.logger.Error("Failed to list due income positions", "error", err)
		return
	}

	if len(positions) == 0 {
		return
	}

	j.logger.Info("Found due positions for income processing", "count", len(positions))

	for _, p := range positions {
		if err := j.processPositionIncome(ctx, p); err != nil {
			j.logger.Error("Failed to process income for position",
				"position_id", p.ID,
				"tenant_id", p.TenantID,
				"error", err)
		}
	}
}

// processPositionIncome creates an income event for the position and updates its next_income_at.
func (j *IncomeSchedulerJob) processPositionIncome(ctx context.Context, p domain.Position) error {
	if p.NextIncomeAt == nil {
		return nil
	}

	amountCents := computeIncomeAmount(p)

	// Create income event
	input := domain.CreatePositionIncomeEventInput{
		DueAt:       *p.NextIncomeAt,
		PositionID:  p.ID,
		AccountID:   p.AccountID,
		IncomeType:  p.IncomeType,
		Currency:    p.Currency,
		AmountCents: amountCents,
	}

	_, err := j.incomeRepo.Create(ctx, p.TenantID, input)
	if err != nil {
		return fmt.Errorf("failed to create income event: %w", err)
	}

	// Update position's next_income_at
	intervalDays := 30 // Default to 30 if nil
	if p.IncomeIntervalDays != nil {
		intervalDays = *p.IncomeIntervalDays
	}

	nextIncomeAt := p.NextIncomeAt.AddDate(0, 0, intervalDays)
	updateInput := domain.UpdatePositionInput{
		NextIncomeAt: &nextIncomeAt,
	}

	_, err = j.positionRepo.Update(ctx, p.TenantID, p.ID, updateInput)
	if err != nil {
		return fmt.Errorf("failed to update position next_income_at: %w", err)
	}

	j.logger.Info("Processed income event for position",
		"position_id", p.ID,
		"tenant_id", p.TenantID,
		"amount_cents", amountCents,
		"new_next_income_at", nextIncomeAt)

	return nil
}

// computeIncomeAmount calculates the income amount in cents for one event.
// Fixed:     income_amount_cents (directly)
// Rate:      ROUND(last_price_cents * quantity * income_rate_bps / 10000)
// Hybrid:    fixed + rate (both non-nil)
func computeIncomeAmount(p domain.Position) int64 {
	var amount int64

	// Fixed part
	if p.IncomeAmountCents != nil {
		amount += *p.IncomeAmountCents
	}

	// Rate part
	if p.IncomeRateBps != nil {
		qty, _ := p.Quantity.Float64()
		ratePart := float64(p.LastPriceCents) * qty * float64(*p.IncomeRateBps) / 10000.0
		amount += int64(math.Round(ratePart))
	}

	return amount
}

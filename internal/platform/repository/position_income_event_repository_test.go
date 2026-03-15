package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPositionIncomeEventRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "01H2PJZ6X7R8K9M4VQB5A7Y2Z1"
	positionID := "01H2PJZ6X7R8K9M4VQB5A7Y2Z2"
	now := time.Now().UTC()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		input := domain.CreatePositionIncomeEventInput{
			PositionID:  positionID,
			IncomeType:  domain.IncomeTypeDividend,
			AmountCents: 1000,
			Currency:    "BRL",
			DueAt:       now,
		}

		mQuerier.On("CreatePositionIncomeEvent", ctx, mock.MatchedBy(func(arg sqlc.CreatePositionIncomeEventParams) bool {
			return arg.TenantID == tenantID &&
				arg.PositionID == input.PositionID &&
				arg.IncomeType == sqlc.IncomeType(input.IncomeType) &&
				arg.AmountCents == input.AmountCents &&
				arg.Currency == input.Currency &&
				arg.EventDate.Time.Equal(input.DueAt)
		})).Return(sqlc.PositionIncomeEvent{
			ID:          "EVENT_1",
			TenantID:    tenantID,
			PositionID:  positionID,
			IncomeType:  sqlc.IncomeTypeDividend,
			AmountCents: 1000,
			Currency:    "BRL",
			EventDate:   pgtype.Date{Time: input.DueAt, Valid: true},
			Status:      sqlc.ReceivableStatusPending,
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

		event, err := repo.Create(ctx, tenantID, input)
		require.NoError(t, err)
		require.NotNil(t, event)
		assert.Equal(t, "EVENT_1", event.ID)
		assert.Equal(t, domain.ReceivableStatusPending, event.Status)
	})

	t.Run("Create Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		mQuerier.On("CreatePositionIncomeEvent", ctx, mock.Anything).
			Return(sqlc.PositionIncomeEvent{}, fmt.Errorf("db error"))

		event, err := repo.Create(ctx, tenantID, domain.CreatePositionIncomeEventInput{})
		require.Error(t, err)
		assert.Nil(t, event)
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		id := "EVENT_1"
		mQuerier.On("GetPositionIncomeEventByID", ctx, sqlc.GetPositionIncomeEventByIDParams{
			ID:       id,
			TenantID: tenantID,
		}).Return(sqlc.PositionIncomeEvent{
			ID:         id,
			TenantID:   tenantID,
			PositionID: positionID,
			Status:     sqlc.ReceivableStatusPending,
		}, nil)

		event, err := repo.GetByID(ctx, tenantID, id)
		require.NoError(t, err)
		require.NotNil(t, event)
		assert.Equal(t, id, event.ID)
	})

	t.Run("GetByID Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		mQuerier.On("GetPositionIncomeEventByID", ctx, mock.Anything).
			Return(sqlc.PositionIncomeEvent{}, fmt.Errorf("db error"))

		event, err := repo.GetByID(ctx, tenantID, "any")
		require.Error(t, err)
		assert.Nil(t, event)
	})

	t.Run("ListByTenant", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		mQuerier.On("ListPositionIncomeEventsByTenant", ctx, tenantID).Return([]sqlc.PositionIncomeEvent{
			{ID: "E1", TenantID: tenantID},
			{ID: "E2", TenantID: tenantID},
		}, nil)

		list, err := repo.ListByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("ListPending", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		mQuerier.On("ListPendingIncomeEvents", ctx, tenantID).Return([]sqlc.PositionIncomeEvent{
			{ID: "E1", TenantID: tenantID, Status: sqlc.ReceivableStatusPending},
		}, nil)

		list, err := repo.ListPending(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("UpdateStatus Success", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		id := "EVENT_1"
		receivedAt := now

		// Mock Get to check transition
		mQuerier.On("GetPositionIncomeEventByID", ctx, sqlc.GetPositionIncomeEventByIDParams{
			ID:       id,
			TenantID: tenantID,
		}).Return(sqlc.PositionIncomeEvent{
			ID:     id,
			Status: sqlc.ReceivableStatusPending,
		}, nil)

		// Mock Update
		mQuerier.On("UpdateIncomeEventStatus", ctx, sqlc.UpdateIncomeEventStatusParams{
			ID:         id,
			TenantID:   tenantID,
			Status:     sqlc.ReceivableStatusReceived,
			RealizedAt: pgtype.Timestamptz{Time: receivedAt, Valid: true},
		}).Return(sqlc.PositionIncomeEvent{
			ID:         id,
			Status:     sqlc.ReceivableStatusReceived,
			RealizedAt: pgtype.Timestamptz{Time: receivedAt, Valid: true},
		}, nil)

		event, err := repo.UpdateStatus(ctx, tenantID, id, domain.ReceivableStatusReceived, &receivedAt)
		require.NoError(t, err)
		require.NotNil(t, event)
		assert.Equal(t, domain.ReceivableStatusReceived, event.Status)
		assert.NotNil(t, event.ReceivedAt)
	})

	t.Run("UpdateStatus Conflict", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		id := "EVENT_1"
		mQuerier.On("GetPositionIncomeEventByID", ctx, mock.Anything).Return(sqlc.PositionIncomeEvent{
			ID:     id,
			Status: sqlc.ReceivableStatusReceived, // Already received
		}, nil)

		event, err := repo.UpdateStatus(ctx, tenantID, id, domain.ReceivableStatusReceived, nil)
		require.ErrorIs(t, err, domain.ErrConflict)
		assert.Nil(t, event)
	})

	t.Run("UpdateStatus Not Found", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		mQuerier.On("GetPositionIncomeEventByID", ctx, mock.Anything).Return(sqlc.PositionIncomeEvent{}, fmt.Errorf("not found"))

		event, err := repo.UpdateStatus(ctx, tenantID, "any", domain.ReceivableStatusReceived, nil)
		require.Error(t, err)
		assert.Nil(t, event)
	})

	t.Run("UpdateStatus Success - Nil Time", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		id := "EVENT_1"

		// Mock Get to check transition
		mQuerier.On("GetPositionIncomeEventByID", ctx, mock.Anything).Return(sqlc.PositionIncomeEvent{
			ID:     id,
			Status: sqlc.ReceivableStatusPending,
		}, nil)

		// Mock Update with nil realized_at
		mQuerier.On("UpdateIncomeEventStatus", ctx, sqlc.UpdateIncomeEventStatusParams{
			ID:         id,
			TenantID:   tenantID,
			Status:     sqlc.ReceivableStatusCancelled,
			RealizedAt: pgtype.Timestamptz{Valid: false},
		}).Return(sqlc.PositionIncomeEvent{
			ID:         id,
			Status:     sqlc.ReceivableStatusCancelled,
			RealizedAt: pgtype.Timestamptz{Valid: false},
		}, nil)

		event, err := repo.UpdateStatus(ctx, tenantID, id, domain.ReceivableStatusCancelled, nil)
		require.NoError(t, err)
		require.NotNil(t, event)
		assert.Equal(t, domain.ReceivableStatusCancelled, event.Status)
		assert.Nil(t, event.ReceivedAt)
	})

	t.Run("UpdateStatus Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		id := "EVENT_1"

		// Mock Get success
		mQuerier.On("GetPositionIncomeEventByID", ctx, mock.Anything).Return(sqlc.PositionIncomeEvent{
			ID:     id,
			Status: sqlc.ReceivableStatusPending,
		}, nil)

		// Mock Update fail
		mQuerier.On("UpdateIncomeEventStatus", ctx, mock.Anything).
			Return(sqlc.PositionIncomeEvent{}, fmt.Errorf("db update error"))

		event, err := repo.UpdateStatus(ctx, tenantID, id, domain.ReceivableStatusReceived, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update income event status")
		assert.Nil(t, event)
	})

	t.Run("ListByTenant Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		mQuerier.On("ListPositionIncomeEventsByTenant", ctx, tenantID).
			Return(nil, fmt.Errorf("db error"))

		list, err := repo.ListByTenant(ctx, tenantID)
		require.Error(t, err)
		assert.Nil(t, list)
	})

	t.Run("ListPending Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionIncomeEventRepository(mQuerier)

		mQuerier.On("ListPendingIncomeEvents", ctx, tenantID).
			Return(nil, fmt.Errorf("db error"))

		list, err := repo.ListPending(ctx, tenantID)
		require.Error(t, err)
		assert.Nil(t, list)
	})
}

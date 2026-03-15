package mocks

import (
	"context"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// PositionRepository is a mock implementation of domain.PositionRepository.
type PositionRepository struct {
	mock.Mock
}

var _ domain.PositionRepository = (*PositionRepository)(nil)

func (m *PositionRepository) Create(ctx context.Context, tenantID string, in domain.CreatePositionInput) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Position), args.Error(1)
}

func (m *PositionRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Position), args.Error(1)
}

func (m *PositionRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Position, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.Position
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Position)
	}
	return r0, args.Error(1)
}

func (m *PositionRepository) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.Position, error) {
	args := m.Called(ctx, tenantID, accountID)
	var r0 []domain.Position
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Position)
	}
	return r0, args.Error(1)
}

func (m *PositionRepository) ListDueIncome(ctx context.Context, before time.Time) ([]domain.Position, error) {
	args := m.Called(ctx, before)
	var r0 []domain.Position
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Position)
	}
	return r0, args.Error(1)
}

func (m *PositionRepository) Update(ctx context.Context, tenantID, id string, in domain.UpdatePositionInput) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, id, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Position), args.Error(1)
}

func (m *PositionRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// PositionSnapshotRepository is a mock implementation of domain.PositionSnapshotRepository.
type PositionSnapshotRepository struct {
	mock.Mock
}

var _ domain.PositionSnapshotRepository = (*PositionSnapshotRepository)(nil)

func (m *PositionSnapshotRepository) Create(ctx context.Context, tenantID string, in domain.CreatePositionSnapshotInput) (*domain.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PositionSnapshot), args.Error(1)
}

func (m *PositionSnapshotRepository) ListByPosition(ctx context.Context, tenantID, positionID string) ([]domain.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, positionID)
	var r0 []domain.PositionSnapshot
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.PositionSnapshot)
	}
	return r0, args.Error(1)
}

func (m *PositionSnapshotRepository) ListByTenantSince(ctx context.Context, tenantID string, since time.Time) ([]domain.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, since)
	var r0 []domain.PositionSnapshot
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.PositionSnapshot)
	}
	return r0, args.Error(1)
}

// PositionIncomeEventRepository is a mock implementation of domain.PositionIncomeEventRepository.
type PositionIncomeEventRepository struct {
	mock.Mock
}

var _ domain.PositionIncomeEventRepository = (*PositionIncomeEventRepository)(nil)

func (m *PositionIncomeEventRepository) Create(ctx context.Context, tenantID string, in domain.CreatePositionIncomeEventInput) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PositionIncomeEvent), args.Error(1)
}

func (m *PositionIncomeEventRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PositionIncomeEvent), args.Error(1)
}

func (m *PositionIncomeEventRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.PositionIncomeEvent
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.PositionIncomeEvent)
	}
	return r0, args.Error(1)
}

func (m *PositionIncomeEventRepository) ListPending(ctx context.Context, tenantID string) ([]domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.PositionIncomeEvent
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.PositionIncomeEvent)
	}
	return r0, args.Error(1)
}

func (m *PositionIncomeEventRepository) UpdateStatus(ctx context.Context, tenantID, id string, status domain.ReceivableStatus, receivedAt *time.Time) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, id, status, receivedAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PositionIncomeEvent), args.Error(1)
}

// PortfolioSnapshotRepository is a mock implementation of domain.PortfolioSnapshotRepository.
type PortfolioSnapshotRepository struct {
	mock.Mock
}

var _ domain.PortfolioSnapshotRepository = (*PortfolioSnapshotRepository)(nil)

func (m *PortfolioSnapshotRepository) Create(ctx context.Context, tenantID string, in domain.CreatePortfolioSnapshotInput) (*domain.PortfolioSnapshot, error) {
	args := m.Called(ctx, tenantID, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PortfolioSnapshot), args.Error(1)
}

func (m *PortfolioSnapshotRepository) GetByDate(ctx context.Context, tenantID string, date time.Time) (*domain.PortfolioSnapshot, error) {
	args := m.Called(ctx, tenantID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PortfolioSnapshot), args.Error(1)
}

func (m *PortfolioSnapshotRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.PortfolioSnapshot, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.PortfolioSnapshot
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.PortfolioSnapshot)
	}
	return r0, args.Error(1)
}

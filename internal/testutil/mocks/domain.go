package mocks

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// AuthRepository is a testify/mock implementation of domain.AuthRepository.
type AuthRepository struct {
	mock.Mock
}

func (m *AuthRepository) CreateOTPRequest(ctx context.Context, input domain.CreateOTPRequestInput) (*domain.OTPRequest, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CreateOTPRequest: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.OTPRequest), err
}

func (m *AuthRepository) GetActiveOTPRequest(ctx context.Context, email string) (*domain.OTPRequest, error) {
	args := m.Called(ctx, email)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetActiveOTPRequest: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.OTPRequest), err
}

func (m *AuthRepository) MarkOTPUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock MarkOTPUsed: %w", e)
	}
	return nil
}

func (m *AuthRepository) DeleteExpiredOTPRequests(ctx context.Context) error {
	args := m.Called(ctx)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock DeleteExpiredOTPRequests: %w", e)
	}
	return nil
}

var _ domain.AuthRepository = (*AuthRepository)(nil)

// UserRepository is a testify/mock implementation of domain.UserRepository.
type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err
}

func (m *UserRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.User, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err
}

func (m *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.GetByEmail: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err
}

func (m *UserRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.User, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.User), err
}

func (m *UserRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateUserInput) (*domain.User, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err
}

func (m *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock UserRepository.UpdateLastLogin: %w", e)
	}
	return nil
}

func (m *UserRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock UserRepository.Delete: %w", e)
	}
	return nil
}

var _ domain.UserRepository = (*UserRepository)(nil)

// AuditRepository is a testify/mock implementation of domain.AuditRepository.
type AuditRepository struct {
	mock.Mock
}

func (m *AuditRepository) Create(ctx context.Context, input domain.CreateAuditLogInput) (*domain.AuditLog, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AuditRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.AuditLog), err
}

func (m *AuditRepository) ListByTenant(ctx context.Context, tenantID string, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, tenantID, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AuditRepository.ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.AuditLog), err
}

func (m *AuditRepository) ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]domain.AuditLog, error) {
	args := m.Called(ctx, tenantID, entityType, entityID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AuditRepository.ListByEntity: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.AuditLog), err
}

var _ domain.AuditRepository = (*AuditRepository)(nil)

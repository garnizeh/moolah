package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type tenantRepo struct {
	q sqlc.Querier
}

// NewTenantRepository creates a new concrete implementation of domain.TenantRepository.
func NewTenantRepository(q sqlc.Querier) domain.TenantRepository {
	return &tenantRepo{q: q}
}

// Create persists a new tenant.
func (r *tenantRepo) Create(ctx context.Context, input domain.CreateTenantInput) (*domain.Tenant, error) {
	arg := sqlc.CreateTenantParams{
		ID:   ulid.New(),
		Name: input.Name,
		Plan: sqlc.TenantPlan(domain.TenantPlanFree),
	}

	t, err := r.q.CreateTenant(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: tenant name already exists", domain.ErrConflict)
		}
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return r.mapTenant(t), nil
}

// GetByID retrieves a tenant by its unique identifier.
func (r *tenantRepo) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	t, err := r.q.GetTenantByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by id: %w", err)
	}

	return r.mapTenant(t), nil
}

// List returns all active (non-deleted) tenants.
func (r *tenantRepo) List(ctx context.Context) ([]domain.Tenant, error) {
	tenants, err := r.q.ListTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	result := make([]domain.Tenant, len(tenants))
	for i, t := range tenants {
		result[i] = *r.mapTenant(t)
	}

	return result, nil
}

// Update modifies an existing tenant's attributes.
func (r *tenantRepo) Update(ctx context.Context, id string, input domain.UpdateTenantInput) (*domain.Tenant, error) {
	// First get current state to handle partial updates
	current, err := r.q.GetTenantByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tenant for update: %w", err)
	}

	arg := sqlc.UpdateTenantParams{
		ID:   id,
		Name: current.Name,
		Plan: current.Plan,
	}

	if input.Name != nil {
		arg.Name = *input.Name
	}
	if input.Plan != nil {
		arg.Plan = sqlc.TenantPlan(*input.Plan)
	}

	t, err := r.q.UpdateTenant(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: tenant name already exists", domain.ErrConflict)
		}
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return r.mapTenant(t), nil
}

// Delete performs a soft-delete on the tenant.
func (r *tenantRepo) Delete(ctx context.Context, id string) error {
	err := r.q.SoftDeleteTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete tenant: %w", err)
	}

	return nil
}

func (r *tenantRepo) mapTenant(t sqlc.Tenant) *domain.Tenant {
	var deletedAt *time.Time
	if t.DeletedAt.Valid {
		deletedAt = &t.DeletedAt.Time
	}

	// Reorder fields to match fieldalignment (domain.Tenant: CreatedAt UpdateAt DeletedAt ID Name Plan)
	return &domain.Tenant{
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
		DeletedAt: deletedAt,
		ID:        t.ID,
		Name:      t.Name,
		Plan:      domain.TenantPlan(t.Plan),
	}
}

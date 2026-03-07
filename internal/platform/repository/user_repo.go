package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type userRepo struct {
	q sqlc.Querier
}

// NewUserRepository creates a new concrete implementation of domain.UserRepository.
func NewUserRepository(q sqlc.Querier) domain.UserRepository {
	return &userRepo{q: q}
}

// Create persists a new user within a tenant.
func (r *userRepo) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	arg := sqlc.CreateUserParams{
		ID:       ulid.New(),
		TenantID: input.TenantID,
		Email:    input.Email,
		Name:     input.Name,
		Role:     sqlc.UserRole(input.Role),
	}

	u, err := r.q.CreateUser(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: user email already exists", domain.ErrConflict)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return r.mapUser(u), nil
}

// GetByID retrieves a user by ID, scoped to a specific tenant.
func (r *userRepo) GetByID(ctx context.Context, tenantID, id string) (*domain.User, error) {
	u, err := r.q.GetUserByID(ctx, sqlc.GetUserByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return r.mapUser(u), nil
}

// GetByEmail retrieves a user by their unique email across all tenants.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.q.GetUserByEmail(ctx, sqlc.GetUserByEmailParams{
		Email: email,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return r.mapUser(u), nil
}

// ListByTenant returns all active users belonging to a household.
func (r *userRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.User, error) {
	users, err := r.q.ListUsersByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	domainUsers := make([]domain.User, len(users))
	for i, u := range users {
		domainUsers[i] = *r.mapUser(u)
	}

	return domainUsers, nil
}

// Update modifies a user's attributes within their tenant.
func (r *userRepo) Update(ctx context.Context, tenantID, id string, input domain.UpdateUserInput) (*domain.User, error) {
	current, err := r.q.GetUserByID(ctx, sqlc.GetUserByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}

	arg := sqlc.UpdateUserParams{
		ID:       id,
		TenantID: tenantID,
		Name:     current.Name,
		Role:     current.Role,
	}

	if input.Name != nil {
		arg.Name = *input.Name
	}
	if input.Role != nil {
		arg.Role = sqlc.UserRole(*input.Role)
	}

	u, err := r.q.UpdateUser(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return r.mapUser(u), nil
}

// UpdateLastLogin updates the login timestamp for a user.
func (r *userRepo) UpdateLastLogin(ctx context.Context, id string) error {
	err := r.q.UpdateUserLastLogin(ctx, sqlc.UpdateUserLastLoginParams{
		ID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update user last login: %w", err)
	}
	return nil
}

// Delete performs a soft-delete on a user.
func (r *userRepo) Delete(ctx context.Context, tenantID, id string) error {
	err := r.q.SoftDeleteUser(ctx, sqlc.SoftDeleteUserParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}
	return nil
}

func (r *userRepo) mapUser(u sqlc.User) *domain.User {
	user := &domain.User{
		ID:        u.ID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      domain.Role(u.Role),
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}

	if u.DeletedAt.Valid {
		user.DeletedAt = &u.DeletedAt.Time
	}
	if u.LastLoginAt.Valid {
		user.LastLoginAt = &u.LastLoginAt.Time
	}

	return user
}

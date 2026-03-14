package repository

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
)

// accountRepo implements the domain.AccountRepository interface using a sqlc.Querier for database interactions.
type accountRepo struct {
	q sqlc.Querier
}

// NewAccountRepository creates a new concrete implementation of domain.AccountRepository.
func NewAccountRepository(q sqlc.Querier) domain.AccountRepository {
	return &accountRepo{q: q}
}

// Create persists a new account for the specified tenant.
func (r *accountRepo) Create(ctx context.Context, tenantID string, input domain.CreateAccountInput) (*domain.Account, error) {
	row, err := r.q.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:           ulid.New(),
		TenantID:     tenantID,
		UserID:       input.UserID,
		Name:         input.Name,
		Type:         sqlc.AccountType(input.Type),
		Currency:     input.Currency,
		BalanceCents: input.InitialCents,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", TranslateError(err))
	}

	return mapAccount(row), nil
}

// GetByID retrieves a specific account by its ID and tenant ID.
func (r *accountRepo) GetByID(ctx context.Context, tenantID, id string) (*domain.Account, error) {
	row, err := r.q.GetAccountByID(ctx, sqlc.GetAccountByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", TranslateError(err))
	}

	return mapAccount(row), nil
}

// ListByTenant returns all active accounts for the given tenant.
func (r *accountRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Account, error) {
	rows, err := r.q.ListAccountsByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", TranslateError(err))
	}

	accounts := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, *mapAccount(row))
	}

	return accounts, nil
}

// ListByUser returns all accounts associated with a specific user within a tenant.
func (r *accountRepo) ListByUser(ctx context.Context, tenantID, userID string) ([]domain.Account, error) {
	rows, err := r.q.ListAccountsByUser(ctx, sqlc.ListAccountsByUserParams{
		TenantID: tenantID,
		UserID:   userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list user accounts: %w", TranslateError(err))
	}

	accounts := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, *mapAccount(row))
	}

	return accounts, nil
}

// Update modifies an existing account's metadata.
func (r *accountRepo) Update(ctx context.Context, tenantID, id string, input domain.UpdateAccountInput) (*domain.Account, error) {
	// First get current account to handle partial updates
	current, err := r.q.GetAccountByID(ctx, sqlc.GetAccountByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account for update: %w", TranslateError(err))
	}

	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}

	accType := current.Type
	// Note: domain.UpdateAccountInput doesn't have Type in the template,
	// but the domain doc says "Update modifies an existing account's attributes".
	// The sqlc query for UpdateAccount includes type.
	// We'll stick to what the input provides.

	currency := current.Currency
	if input.Currency != nil {
		currency = *input.Currency
	}

	row, err := r.q.UpdateAccount(ctx, sqlc.UpdateAccountParams{
		TenantID: tenantID,
		ID:       id,
		Name:     name,
		Type:     accType,
		Currency: currency,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", TranslateError(err))
	}

	return mapAccount(row), nil
}

// UpdateBalance updates the balance of an account.
func (r *accountRepo) UpdateBalance(ctx context.Context, tenantID, id string, newBalanceCents int64) error {
	err := r.q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
		TenantID:     tenantID,
		ID:           id,
		BalanceCents: newBalanceCents,
	})
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", TranslateError(err))
	}

	return nil
}

// Delete performs a soft delete on the specified account.
func (r *accountRepo) Delete(ctx context.Context, tenantID, id string) error {
	err := r.q.SoftDeleteAccount(ctx, sqlc.SoftDeleteAccountParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", TranslateError(err))
	}

	return nil
}

// CreateOTPRequest creates a new OTP request for the given email, code hash, and expiration time.
func mapAccount(row sqlc.Account) *domain.Account {
	return &domain.Account{
		ID:           row.ID,
		TenantID:     row.TenantID,
		UserID:       row.UserID,
		Name:         row.Name,
		Type:         domain.AccountType(row.Type),
		Currency:     row.Currency,
		BalanceCents: row.BalanceCents,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		DeletedAt:    &row.DeletedAt.Time, // Note: sqlc might use pgtype.Timestamp for nulls
	}
}

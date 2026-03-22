package postgres

import (
	"context"
	"encoding/json"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/conv"
)

// Repository implements domain repository interfaces using sqlc.
type Repository struct {
	q *Queries
}

// NewRepository creates a new Postgres repository.
func NewRepository(db DBTX) *Repository {
	return &Repository{
		q: New(db),
	}
}

// Ensure Repository implements the domain interfaces.
var (
	_ domain.CurrencyRepository = (*Repository)(nil)
	_ domain.EntityRepository   = (*Repository)(nil)
	_ domain.AccountRepository  = (*Repository)(nil)
)

// --- CurrencyRepository ---

func (r *Repository) CreateCurrency(ctx context.Context, c *domain.Currency) error {
	decimals, err := conv.ToInt32(c.FallbackDecimals)
	if err != nil {
		return err
	}
	arg := CreateCurrencyParams{
		ID:               c.ID,
		Code:             c.Code,
		Symbol:           c.Symbol,
		FallbackDecimals: decimals,
		Config:           c.Config,
	}
	_, err = r.q.CreateCurrency(ctx, arg)
	return err
}

func (r *Repository) GetCurrencyByID(ctx context.Context, id string) (*domain.Currency, error) {
	row, err := r.q.GetCurrency(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainCurrency(row), nil
}

func (r *Repository) ListCurrencies(ctx context.Context) ([]*domain.Currency, error) {
	rows, err := r.q.ListCurrencies(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Currency, len(rows))
	for i, row := range rows {
		res[i] = toDomainCurrency(row)
	}
	return res, nil
}

func (r *Repository) UpdateCurrency(ctx context.Context, c *domain.Currency) error {
	decimals, err := conv.ToInt32(c.FallbackDecimals)
	if err != nil {
		return err
	}
	arg := UpdateCurrencyParams{
		ID:               c.ID,
		Code:             c.Code,
		Symbol:           c.Symbol,
		FallbackDecimals: decimals,
		Config:           c.Config,
	}
	_, err = r.q.UpdateCurrency(ctx, arg)
	return err
}

// --- EntityRepository ---

func (r *Repository) CreateEntity(ctx context.Context, e *domain.Entity) error {
	arg := CreateEntityParams{
		ID:       e.ID,
		Name:     e.Name,
		Role:     e.Role,
		Metadata: e.Metadata,
	}
	_, err := r.q.CreateEntity(ctx, arg)
	return err
}

func (r *Repository) GetEntityByID(ctx context.Context, id string) (*domain.Entity, error) {
	row, err := r.q.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainEntity(row), nil
}

func (r *Repository) ListEntities(ctx context.Context) ([]*domain.Entity, error) {
	rows, err := r.q.ListEntities(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Entity, len(rows))
	for i, row := range rows {
		res[i] = toDomainEntity(row)
	}
	return res, nil
}

func (r *Repository) UpdateEntity(ctx context.Context, e *domain.Entity) error {
	arg := UpdateEntityParams{
		ID:       e.ID,
		Name:     e.Name,
		Role:     e.Role,
		Metadata: e.Metadata,
	}
	_, err := r.q.UpdateEntity(ctx, arg)
	return err
}

func (r *Repository) DeleteEntity(ctx context.Context, id string) error {
	return r.q.DeleteEntity(ctx, id)
}

// --- AccountRepository ---

func (r *Repository) CreateAccount(ctx context.Context, a *domain.Account) error {
	arg := CreateAccountParams{
		ID:           a.ID,
		EntityID:     a.EntityID,
		CurrencyID:   a.CurrencyID,
		Name:         a.Name,
		Type:         a.Type,
		BalanceCents: a.BalanceCents,
		Metadata:     a.Metadata,
	}
	_, err := r.q.CreateAccount(ctx, arg)
	return err
}

func (r *Repository) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	row, err := r.q.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainAccount(row), nil
}

func (r *Repository) ListAccountsByEntity(ctx context.Context, entityID string) ([]*domain.Account, error) {
	rows, err := r.q.ListAccountsByEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Account, len(rows))
	for i, row := range rows {
		res[i] = toDomainAccount(row)
	}
	return res, nil
}

func (r *Repository) UpdateAccountBalance(ctx context.Context, id string, balanceCents int64) (*domain.Account, error) {
	arg := UpdateAccountBalanceParams{
		ID:           id,
		BalanceCents: balanceCents,
	}
	row, err := r.q.UpdateAccountBalance(ctx, arg)
	if err != nil {
		return nil, err
	}
	return toDomainAccount(row), nil
}

func (r *Repository) DeleteAccount(ctx context.Context, id string) error {
	return r.q.DeleteAccount(ctx, id)
}

// --- Helpers ---

func toDomainCurrency(row Currency) *domain.Currency {
	return &domain.Currency{
		ID:               row.ID,
		Code:             row.Code,
		Symbol:           row.Symbol,
		FallbackDecimals: int(row.FallbackDecimals),
		Config:           json.RawMessage(row.Config),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func toDomainEntity(row Entity) *domain.Entity {
	e := &domain.Entity{
		ID:        row.ID,
		Name:      row.Name,
		Role:      row.Role,
		Metadata:  json.RawMessage(row.Metadata),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
	if row.DeletedAt.Valid {
		e.DeletedAt = &row.DeletedAt.Time
	}
	return e
}

func toDomainAccount(row Account) *domain.Account {
	a := &domain.Account{
		ID:           row.ID,
		EntityID:     row.EntityID,
		CurrencyID:   row.CurrencyID,
		Name:         row.Name,
		Type:         row.Type,
		BalanceCents: row.BalanceCents,
		Metadata:     json.RawMessage(row.Metadata),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
	if row.DeletedAt.Valid {
		a.DeletedAt = &row.DeletedAt.Time
	}
	return a
}

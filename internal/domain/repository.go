package domain

import "context"

// CurrencyRepository defines the storage interface for currencies.
type CurrencyRepository interface {
	CreateCurrency(ctx context.Context, c *Currency) error
	GetCurrencyByID(ctx context.Context, id string) (*Currency, error)
	ListCurrencies(ctx context.Context) ([]*Currency, error)
	UpdateCurrency(ctx context.Context, c *Currency) error
}

// EntityRepository defines the storage interface for entities.
type EntityRepository interface {
	CreateEntity(ctx context.Context, e *Entity) error
	GetEntityByID(ctx context.Context, id string) (*Entity, error)
	ListEntities(ctx context.Context) ([]*Entity, error)
	UpdateEntity(ctx context.Context, e *Entity) error
	DeleteEntity(ctx context.Context, id string) error
}

// AccountRepository defines the storage interface for accounts.
type AccountRepository interface {
	CreateAccount(ctx context.Context, a *Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccountsByEntity(ctx context.Context, entityID string) ([]*Account, error)
	UpdateAccountBalance(ctx context.Context, id string, balanceCents int64) (*Account, error)

	DeleteAccount(ctx context.Context, id string) error
}

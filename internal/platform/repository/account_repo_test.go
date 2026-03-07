package repository

import (
"context"
"errors"
"testing"
"time"

"github.com/garnizeh/moolah/internal/domain"
"github.com/garnizeh/moolah/internal/platform/db/sqlc"
"github.com/garnizeh/moolah/internal/testutil/mocks"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgconn"
"github.com/jackc/pgx/v5/pgtype"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"
"github.com/stretchr/testify/require"
)

func TestAccountRepo_Create(t *testing.T) {
t.Parallel()

ctx := context.Background()
tenantID := "01H7XFRP9K1A1A1A1A1A1A1A1A"
input := domain.CreateAccountInput{
UserID:       "01H7XFRP9K1A1A1A1A1A1A1A1B",
Name:         "Main Checking",
Type:         domain.AccountTypeChecking,
Currency:     "USD",
InitialCents: 1000,
}

t.Run("success", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("CreateAccount", ctx, mock.MatchedBy(func(arg sqlc.CreateAccountParams) bool {
return arg.TenantID == tenantID &&
arg.UserID == input.UserID &&
arg.Name == input.Name &&
arg.Type == sqlc.AccountType(input.Type) &&
arg.Currency == input.Currency &&
arg.BalanceCents == input.InitialCents &&
arg.ID != ""
})).Return(sqlc.Account{
ID:           "01H7XFRP9K1A1A1A1A1A1A1A1C",
TenantID:     tenantID,
UserID:       input.UserID,
Name:         input.Name,
Type:         sqlc.AccountType(input.Type),
Currency:     input.Currency,
BalanceCents: input.InitialCents,
CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
UpdatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
}, nil)

account, err := repo.Create(ctx, tenantID, input)

require.NoError(t, err)
assert.NotNil(t, account)
assert.Equal(t, input.Name, account.Name)
assert.Equal(t, input.InitialCents, account.BalanceCents)
mockQuerier.AssertExpectations(t)
})

t.Run("duplicate name", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("CreateAccount", ctx, mock.Anything).Return(sqlc.Account{}, &pgconn.PgError{Code: "23505"})

account, err := repo.Create(ctx, tenantID, input)

require.Error(t, err)
assert.Nil(t, account)
assert.ErrorIs(t, err, domain.ErrConflict)
})
}

func TestAccountRepo_GetByID(t *testing.T) {
t.Parallel()

ctx := context.Background()
tenantID := "01H7XFRP9K1A1A1A1A1A1A1A1A"
accountID := "01H7XFRP9K1A1A1A1A1A1A1A1C"

t.Run("success", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("GetAccountByID", ctx, sqlc.GetAccountByIDParams{
TenantID: tenantID,
ID:       accountID,
}).Return(sqlc.Account{
ID:        accountID,
TenantID:  tenantID,
Name:      "Main Checking",
CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
}, nil)

account, err := repo.GetByID(ctx, tenantID, accountID)

require.NoError(t, err)
assert.NotNil(t, account)
assert.Equal(t, accountID, account.ID)
})

t.Run("not found", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{}, pgx.ErrNoRows)

account, err := repo.GetByID(ctx, tenantID, accountID)

require.Error(t, err)
assert.Nil(t, account)
assert.ErrorIs(t, err, domain.ErrNotFound)
})
}

func TestAccountRepo_ListByTenant(t *testing.T) {
t.Parallel()

ctx := context.Background()
tenantID := "01H7XFRP9K1A1A1A1A1A1A1A1A"

t.Run("success", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("ListAccountsByTenant", ctx, tenantID).Return([]sqlc.Account{
{ID: "acc1", Name: "Account 1", CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
{ID: "acc2", Name: "Account 2", CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
}, nil)

accounts, err := repo.ListByTenant(ctx, tenantID)

require.NoError(t, err)
assert.Len(t, accounts, 2)
assert.Equal(t, "acc1", accounts[0].ID)
assert.Equal(t, "acc2", accounts[1].ID)
})
}

func TestAccountRepo_UpdateBalance(t *testing.T) {
t.Parallel()

ctx := context.Background()
tenantID := "01H7XFRP9K1A1A1A1A1A1A1A1A"
accountID := "01H7XFRP9K1A1A1A1A1A1A1A1C"
newBalance := int64(5000)

t.Run("success", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("UpdateAccountBalance", ctx, sqlc.UpdateAccountBalanceParams{
TenantID:     tenantID,
ID:           accountID,
BalanceCents: newBalance,
}).Return(nil)

err := repo.UpdateBalance(ctx, tenantID, accountID, newBalance)

require.NoError(t, err)
})

t.Run("not found", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("UpdateAccountBalance", ctx, mock.Anything).Return(pgx.ErrNoRows)

err := repo.UpdateBalance(ctx, tenantID, accountID, newBalance)

require.Error(t, err)
assert.ErrorIs(t, err, domain.ErrNotFound)
})
}

func TestAccountRepo_Delete(t *testing.T) {
t.Parallel()

ctx := context.Background()
tenantID := "01H7XFRP9K1A1A1A1A1A1A1A1A"
accountID := "01H7XFRP9K1A1A1A1A1A1A1A1C"

t.Run("success", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("SoftDeleteAccount", ctx, sqlc.SoftDeleteAccountParams{
TenantID: tenantID,
ID:       accountID,
}).Return(nil)

err := repo.Delete(ctx, tenantID, accountID)

require.NoError(t, err)
})

t.Run("error", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("SoftDeleteAccount", ctx, mock.Anything).Return(errors.New("db error"))

err := repo.Delete(ctx, tenantID, accountID)

require.Error(t, err)
assert.Contains(t, err.Error(), "failed to delete account")
})
}

func TestAccountRepo_ListByUser(t *testing.T) {
t.Parallel()

ctx := context.Background()
tenantID := "tenant_id"
userID := "user_id"

t.Run("success", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("ListAccountsByUser", ctx, sqlc.ListAccountsByUserParams{
TenantID: tenantID,
UserID:   userID,
}).Return([]sqlc.Account{
{ID: "acc1", Name: "Account 1"},
}, nil)

accounts, err := repo.ListByUser(ctx, tenantID, userID)

require.NoError(t, err)
assert.Len(t, accounts, 1)
assert.Equal(t, "acc1", accounts[0].ID)
})

t.Run("error", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("ListAccountsByUser", ctx, mock.Anything).Return([]sqlc.Account(nil), errors.New("db error"))

accounts, err := repo.ListByUser(ctx, tenantID, userID)

require.Error(t, err)
assert.Nil(t, accounts)
assert.Contains(t, err.Error(), "failed to list user accounts")
})
}

func TestAccountRepo_Update(t *testing.T) {
t.Parallel()

ctx := context.Background()
tenantID := "tenant_id"
accountID := "acc_id"
newName := "New Name"
input := domain.UpdateAccountInput{
Name: &newName,
}

t.Run("success", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("GetAccountByID", ctx, sqlc.GetAccountByIDParams{
TenantID: tenantID,
ID:       accountID,
}).Return(sqlc.Account{
ID:       accountID,
TenantID: tenantID,
Name:     "Old Name",
Type:     sqlc.AccountTypeChecking,
Currency: "USD",
}, nil)

mockQuerier.On("UpdateAccount", ctx, mock.MatchedBy(func(arg sqlc.UpdateAccountParams) bool {
return arg.ID == accountID && arg.Name == newName && arg.Type == sqlc.AccountTypeChecking
})).Return(sqlc.Account{
ID:   accountID,
Name: newName,
}, nil)

account, err := repo.Update(ctx, tenantID, accountID, input)

require.NoError(t, err)
assert.Equal(t, newName, account.Name)
})

t.Run("update currency only", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

currency := "EUR"
input := domain.UpdateAccountInput{
Currency: &currency,
}

mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{
ID:       accountID,
Name:     "Stay",
Type:     sqlc.AccountTypeSavings,
Currency: "USD",
}, nil)

mockQuerier.On("UpdateAccount", ctx, mock.MatchedBy(func(arg sqlc.UpdateAccountParams) bool {
return arg.Currency == currency && arg.Name == "Stay" && arg.Type == sqlc.AccountTypeSavings
})).Return(sqlc.Account{ID: accountID}, nil)

_, err := repo.Update(ctx, tenantID, accountID, input)
assert.NoError(t, err)
})

t.Run("not found", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{}, pgx.ErrNoRows)

account, err := repo.Update(ctx, tenantID, accountID, input)

require.Error(t, err)
assert.ErrorIs(t, err, domain.ErrNotFound)
assert.Nil(t, account)
})

t.Run("conflict", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{ID: accountID}, nil)
mockQuerier.On("UpdateAccount", ctx, mock.Anything).Return(sqlc.Account{}, &pgconn.PgError{Code: "23505"})

account, err := repo.Update(ctx, tenantID, accountID, input)

require.Error(t, err)
assert.ErrorIs(t, err, domain.ErrConflict)
assert.Nil(t, account)
})

t.Run("generic error during update", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{ID: accountID}, nil)
mockQuerier.On("UpdateAccount", ctx, mock.Anything).Return(sqlc.Account{}, errors.New("db error"))

account, err := repo.Update(ctx, tenantID, accountID, input)

require.Error(t, err)
assert.Contains(t, err.Error(), "failed to update account")
assert.Nil(t, account)
})
}

func TestAccountRepo_GenericErrors(t *testing.T) {
t.Parallel()

ctx := context.Background()
tenantID := "tenant_id"

t.Run("create error", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("CreateAccount", ctx, mock.Anything).Return(sqlc.Account{}, errors.New("db error"))

account, err := repo.Create(ctx, tenantID, domain.CreateAccountInput{})
require.Error(t, err)
assert.Contains(t, err.Error(), "failed to create account")
assert.Nil(t, account)
})

t.Run("get by id error", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{}, errors.New("db error"))

account, err := repo.GetByID(ctx, tenantID, "1")
require.Error(t, err)
assert.Contains(t, err.Error(), "failed to get account")
assert.Nil(t, account)
})

t.Run("list by tenant error", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("ListAccountsByTenant", ctx, mock.Anything).Return([]sqlc.Account(nil), errors.New("db error"))

accounts, err := repo.ListByTenant(ctx, tenantID)
require.Error(t, err)
assert.Contains(t, err.Error(), "failed to list accounts")
assert.Nil(t, accounts)
})

t.Run("update balance generic error", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("UpdateAccountBalance", ctx, mock.Anything).Return(errors.New("db error"))

err := repo.UpdateBalance(ctx, tenantID, "1", 100)
require.Error(t, err)
assert.Contains(t, err.Error(), "failed to update account balance")
})

t.Run("update get account generic error", func(t *testing.T) {
t.Parallel()
mockQuerier := new(mocks.Querier)
repo := NewAccountRepository(mockQuerier)

mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{}, errors.New("db error"))

account, err := repo.Update(ctx, tenantID, "1", domain.UpdateAccountInput{})
require.Error(t, err)
assert.Contains(t, err.Error(), "failed to get account for update")
assert.Nil(t, account)
})
}

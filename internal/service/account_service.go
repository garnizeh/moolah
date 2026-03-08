package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/garnizeh/moolah/internal/domain"
)

type accountService struct {
	accountRepo domain.AccountRepository
	userRepo    domain.UserRepository
	auditRepo   domain.AuditRepository
}

// NewAccountService creates a new instance of AccountService.
func NewAccountService(
	accountRepo domain.AccountRepository,
	userRepo domain.UserRepository,
	auditRepo domain.AuditRepository,
) domain.AccountService {
	return &accountService{
		accountRepo: accountRepo,
		userRepo:    userRepo,
		auditRepo:   auditRepo,
	}
}

func (s *accountService) Create(ctx context.Context, tenantID string, input domain.CreateAccountInput) (*domain.Account, error) {
	// 1. Verify that the user belongs to the tenant.
	user, err := s.userRepo.GetByID(ctx, tenantID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("account service: failed to verify user for account creation: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("account service: user not found: %w", domain.ErrForbidden)
	}

	// 2. Persist account.
	account, err := s.accountRepo.Create(ctx, tenantID, input)
	if err != nil {
		return nil, fmt.Errorf("account service: failed to create account: %w", err)
	}

	// 3. Write audit log.
	newValues, err := json.Marshal(map[string]any{
		"name":          account.Name,
		"type":          account.Type,
		"currency":      account.Currency,
		"balance_cents": account.BalanceCents,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal audit trail for account creation", "error", err)
	}

	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    input.UserID, // Using the owner as actor for now, or could come from context.
		Action:     domain.AuditActionCreate,
		EntityType: "account",
		EntityID:   account.ID,
		ActorRole:  user.Role,
		NewValues:  newValues,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create audit log for account creation", "error", err)
	}

	return account, nil
}

func (s *accountService) GetByID(ctx context.Context, tenantID, id string) (*domain.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("account service: failed to get account: %w", err)
	}
	return account, nil
}

func (s *accountService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Account, error) {
	accounts, err := s.accountRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("account service: failed to list accounts: %w", err)
	}
	return accounts, nil
}

func (s *accountService) ListByUser(ctx context.Context, tenantID, userID string) ([]domain.Account, error) {
	accounts, err := s.accountRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("account service: failed to list user accounts: %w", err)
	}
	return accounts, nil
}

func (s *accountService) Update(ctx context.Context, tenantID, id string, input domain.UpdateAccountInput) (*domain.Account, error) {
	oldAccount, err := s.accountRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("account service: failed to fetch existing account: %w", err)
	}

	account, err := s.accountRepo.Update(ctx, tenantID, id, input)
	if err != nil {
		return nil, fmt.Errorf("account service: failed to update account: %w", err)
	}

	newValuesMap := make(map[string]any)
	oldValuesMap := make(map[string]any)

	if input.Name != nil && *input.Name != oldAccount.Name {
		newValuesMap["name"] = *input.Name
		oldValuesMap["name"] = oldAccount.Name
	}
	if input.Currency != nil && *input.Currency != oldAccount.Currency {
		newValuesMap["currency"] = *input.Currency
		oldValuesMap["currency"] = oldAccount.Currency
	}

	if len(newValuesMap) > 0 {
		oldValues, err := json.Marshal(oldValuesMap)
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal old values for account update audit", "error", err)
		}
		newValues, err := json.Marshal(newValuesMap)
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal new values for account update audit", "error", err)
		}

		_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
			TenantID:   tenantID,
			ActorID:    account.UserID,
			Action:     domain.AuditActionUpdate,
			EntityType: "account",
			EntityID:   id,
			ActorRole:  domain.RoleMember, // Defaulting for now
			OldValues:  oldValues,
			NewValues:  newValues,
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to create audit log for account update", "error", err)
		}
	}

	return account, nil
}

func (s *accountService) Delete(ctx context.Context, tenantID, id string) error {
	account, err := s.accountRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("account service: failed to find account for deletion: %w", err)
	}

	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    account.UserID,
		Action:     domain.AuditActionSoftDelete,
		EntityType: "account",
		EntityID:   id,
		ActorRole:  domain.RoleMember,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create audit log for account deletion", "error", err)
	}

	if err := s.accountRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("account service: failed to delete account: %w", err)
	}

	return nil
}

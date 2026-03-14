package service

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
)

// masterPurchaseService provides business logic for managing master purchases, including creation, retrieval, updating, and deletion of master purchase records. It also includes logic for projecting installment details based on the total amount and installment count.
type masterPurchaseService struct {
	mpRepo      domain.MasterPurchaseRepository
	accountRepo domain.AccountRepository
	catRepo     domain.CategoryRepository
}

// NewMasterPurchaseService creates a new MasterPurchaseService implementation.
func NewMasterPurchaseService(
	mpRepo domain.MasterPurchaseRepository,
	accountRepo domain.AccountRepository,
	catRepo domain.CategoryRepository,
) domain.MasterPurchaseService {
	return &masterPurchaseService{
		mpRepo:      mpRepo,
		accountRepo: accountRepo,
		catRepo:     catRepo,
	}
}

// Create validates the input and creates a new master purchase record. It checks that the referenced account exists and is a credit card, and that the referenced category exists and is an expense category. It returns the created master purchase or an error if validation fails or the record cannot be created.
func (s *masterPurchaseService) Create(ctx context.Context, tenantID string, input domain.CreateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	// 1. Verify account belongs to tenant and is of type credit_card.
	acc, err := s.accountRepo.GetByID(ctx, tenantID, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to verify account: %w", err)
	}
	if acc.Type != domain.AccountTypeCreditCard {
		return nil, fmt.Errorf("master purchase service: account %s is not a credit card: %w", input.AccountID, domain.ErrInvalidInput)
	}

	// 2. Verify category belongs to tenant and is of type expense.
	cat, err := s.catRepo.GetByID(ctx, tenantID, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to verify category: %w", err)
	}
	if cat.Type != domain.CategoryTypeExpense {
		return nil, fmt.Errorf("master purchase service: category %s must be of type expense: %w", input.CategoryID, domain.ErrInvalidInput)
	}

	// 3. Create master purchase.
	mp, err := s.mpRepo.Create(ctx, tenantID, input)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to create record: %w", err)
	}

	return mp, nil
}

// GetByID retrieves a master purchase by its ID and tenant ID. It returns the master purchase or an error if the record cannot be found or retrieved.
func (s *masterPurchaseService) GetByID(ctx context.Context, tenantID, id string) (*domain.MasterPurchase, error) {
	mp, err := s.mpRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to get record: %w", err)
	}
	return mp, nil
}

// ListByTenant returns all master purchases for a given tenant ID. It retrieves the list of master purchases or returns an error if the records cannot be retrieved.
func (s *masterPurchaseService) ListByTenant(ctx context.Context, tenantID string) ([]domain.MasterPurchase, error) {
	mps, err := s.mpRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to list by tenant: %w", err)
	}
	return mps, nil
}

// ListByAccount returns all master purchases for a given tenant ID and account ID. It retrieves the list of master purchases or returns an error if the records cannot be retrieved.
func (s *masterPurchaseService) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.MasterPurchase, error) {
	mps, err := s.mpRepo.ListByAccount(ctx, tenantID, accountID)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to list by account: %w", err)
	}
	return mps, nil
}

// Update validates the input and updates an existing master purchase record. It checks that if the category is being updated, the new category exists and is an expense category. It returns the updated master purchase or an error if validation fails or the record cannot be updated.
func (s *masterPurchaseService) Update(ctx context.Context, tenantID, id string, input domain.UpdateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	// If category is provided, verify it exists and is an expense.
	if input.CategoryID != nil {
		cat, err := s.catRepo.GetByID(ctx, tenantID, *input.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("master purchase service: failed to verify new category: %w", err)
		}
		if cat == nil {
			return nil, fmt.Errorf("master purchase service: category %s not found: %w", *input.CategoryID, domain.ErrNotFound)
		}
		if cat.Type != domain.CategoryTypeExpense {
			return nil, fmt.Errorf("master purchase service: category %s must be of type expense: %w", *input.CategoryID, domain.ErrInvalidInput)
		}
	}

	mp, err := s.mpRepo.Update(ctx, tenantID, id, input)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to update record: %w", err)
	}
	return mp, nil
}

// Delete performs a soft delete of a master purchase record by its ID and tenant ID. It returns an error if the record cannot be deleted.
func (s *masterPurchaseService) Delete(ctx context.Context, tenantID, id string) error {
	err := s.mpRepo.Delete(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("master purchase service: failed to delete record: %w", err)
	}
	return nil
}

// ProjectInstallments computes each instalment's due date and amount.
//
// Remainder-cent rule: when TotalAmountCents is not evenly divisible by
// InstallmentCount, the LAST instalment absorbs all remaining cents so that:
//
//	sum(result[i].AmountCents for i in 0..N-1) == mp.TotalAmountCents
//
// Example: total=1000, count=3 → base=333, remainder=1 → [333, 333, 334]
// Example: total=1001, count=3 → base=333, remainder=2 → [333, 333, 335]
// Example: total=1200, count=3 → base=400, remainder=0 → [400, 400, 400]
func (s *masterPurchaseService) ProjectInstallments(mp *domain.MasterPurchase) []domain.ProjectedInstallment {
	if mp.InstallmentCount <= 0 {
		return nil
	}

	result := make([]domain.ProjectedInstallment, mp.InstallmentCount)
	base := mp.TotalAmountCents / int64(mp.InstallmentCount)
	remainder := mp.TotalAmountCents % int64(mp.InstallmentCount)

	for i := range result {
		amount := base
		if i == int(mp.InstallmentCount)-1 {
			amount += remainder
		}
		result[i] = domain.ProjectedInstallment{
			InstallmentNumber: int32(i + 1),
			DueDate:           mp.FirstInstallmentDate.AddDate(0, i, 0),
			AmountCents:       amount,
		}
	}
	return result
}

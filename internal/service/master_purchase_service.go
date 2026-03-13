package service

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
)

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

func (s *masterPurchaseService) GetByID(ctx context.Context, tenantID, id string) (*domain.MasterPurchase, error) {
	mp, err := s.mpRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to get record: %w", err)
	}
	return mp, nil
}

func (s *masterPurchaseService) ListByTenant(ctx context.Context, tenantID string) ([]domain.MasterPurchase, error) {
	mps, err := s.mpRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to list by tenant: %w", err)
	}
	return mps, nil
}

func (s *masterPurchaseService) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.MasterPurchase, error) {
	mps, err := s.mpRepo.ListByAccount(ctx, tenantID, accountID)
	if err != nil {
		return nil, fmt.Errorf("master purchase service: failed to list by account: %w", err)
	}
	return mps, nil
}

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

func (s *masterPurchaseService) Delete(ctx context.Context, tenantID, id string) error {
	err := s.mpRepo.Delete(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("master purchase service: failed to delete record: %w", err)
	}
	return nil
}

func (s *masterPurchaseService) ProjectInstallments(mp *domain.MasterPurchase) []domain.ProjectedInstallment {
	if mp.InstallmentCount <= 0 {
		return nil
	}

	installments := make([]domain.ProjectedInstallment, 0, mp.InstallmentCount)
	baseAmount := mp.TotalAmountCents / int64(mp.InstallmentCount)
	remainder := mp.TotalAmountCents % int64(mp.InstallmentCount)

	startDate := mp.FirstInstallmentDate

	for i := int32(1); i <= mp.InstallmentCount; i++ {
		amount := baseAmount
		if i == mp.InstallmentCount {
			amount += remainder
		}

		// Calculate due date (simplistic: add i-1 months to start date).
		dueDate := startDate.AddDate(0, int(i-1), 0)

		installments = append(installments, domain.ProjectedInstallment{
			InstallmentNumber: i,
			AmountCents:       amount,
			DueDate:           dueDate,
		})
	}

	return installments
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InvoiceCloser materializes installment transactions at invoice-close time.
type InvoiceCloser struct {
	mpRepo    domain.MasterPurchaseRepository
	txRepo    domain.TransactionRepository
	auditRepo domain.AuditRepository
	accRepo   domain.AccountRepository
	mpSvc     domain.MasterPurchaseService
	db        *pgxpool.Pool
}

// NewInvoiceCloser creates a new InvoiceCloser.
func NewInvoiceCloser(
	mpRepo domain.MasterPurchaseRepository,
	txRepo domain.TransactionRepository,
	auditRepo domain.AuditRepository,
	accRepo domain.AccountRepository,
	mpSvc domain.MasterPurchaseService,
	db *pgxpool.Pool,
) *InvoiceCloser {
	return &InvoiceCloser{
		mpRepo:    mpRepo,
		txRepo:    txRepo,
		auditRepo: auditRepo,
		accRepo:   accRepo,
		mpSvc:     mpSvc,
		db:        db,
	}
}

// CloseInvoice finds all open master purchases for the account due on or before
// closingDate, materializes the current installment as a transaction, and advances
// paid_installments. Runs each master purchase in its own DB transaction.
func (c *InvoiceCloser) CloseInvoice(
	ctx context.Context,
	tenantID string,
	accountID string,
	closingDate time.Time,
) (domain.CloseInvoiceResult, error) {
	result := domain.CloseInvoiceResult{
		Errors: []error{},
	}

	// 1. Validate account exists and is a credit card
	acc, err := c.accRepo.GetByID(ctx, tenantID, accountID)
	if err != nil {
		return result, fmt.Errorf("failed to get account: %w", err)
	}
	if acc.Type != domain.AccountTypeCreditCard {
		return result, fmt.Errorf("%w: account is not a credit card", domain.ErrInvalidInput)
	}

	// 2. List pending master purchases for this account
	mps, err := c.mpRepo.ListByAccount(ctx, tenantID, accountID)
	if err != nil {
		return result, fmt.Errorf("failed to list pending master purchases: %w", err)
	}

	for _, mp := range mps {
		if mp.Status == domain.MasterPurchaseStatusClosed {
			continue
		}

		// Project installments to find the one to materialize
		schedule := c.mpSvc.ProjectInstallments(&mp)
		if int(mp.PaidInstallments) >= len(schedule) {
			slog.WarnContext(ctx, "master purchase already fully paid but not closed",
				slog.String("master_purchase_id", mp.ID),
				slog.Int("paid", int(mp.PaidInstallments)),
				slog.Int("total", len(schedule)))
			continue
		}

		currentInstalment := schedule[mp.PaidInstallments]
		if currentInstalment.DueDate.After(closingDate) {
			continue
		}

		// Materialize installment
		err := c.materializeInstallment(ctx, tenantID, &mp, currentInstalment, acc)
		if err != nil {
			slog.WarnContext(ctx, "failed to materialize installment",
				slog.String("master_purchase_id", mp.ID),
				slog.Int("installment_number", int(currentInstalment.InstallmentNumber)),
				slog.Any("error", err))
			result.Errors = append(result.Errors, fmt.Errorf("master_purchase %s: %w", mp.ID, err))
			continue
		}

		result.ProcessedCount++
	}

	return result, nil
}

func (c *InvoiceCloser) materializeInstallment(
	ctx context.Context,
	tenantID string,
	mp *domain.MasterPurchase,
	installment domain.ProjectedInstallment,
	acc *domain.Account,
) error {
	// Start transaction if pool is available
	var tx pgx.Tx
	var err error
	if c.db != nil {
		tx, err = c.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				rbErr := tx.Rollback(ctx)
				if rbErr != nil {
					slog.ErrorContext(ctx, "failed to rollback transaction", slog.Any("error", rbErr))
				}
			}
		}()
	}

	// 1. Create Transaction
	txInput := domain.CreateTransactionInput{
		OccurredAt:       installment.DueDate,
		AccountID:        mp.AccountID,
		CategoryID:       mp.CategoryID,
		UserID:           mp.UserID,
		Description:      fmt.Sprintf("%s (%d/%d)", mp.Description, installment.InstallmentNumber, mp.InstallmentCount),
		MasterPurchaseID: mp.ID,
		Type:             domain.TransactionTypeExpense,
		AmountCents:      installment.AmountCents,
	}

	createdTx, err := c.txRepo.Create(ctx, tenantID, txInput)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// 2. Audit Log (SYSTEM actor)
	metadata, err := json.Marshal(map[string]any{
		"master_purchase_id": mp.ID,
		"installment":        installment.InstallmentNumber,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal audit metadata: %w", err)
	}

	audioInput := domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    domain.ActorSystem,
		Action:     domain.AuditActionCreate,
		EntityType: "transaction",
		EntityID:   createdTx.ID,
		ActorRole:  domain.RoleAdmin, // SYSTEM acts as Admin
		Metadata:   metadata,
	}
	_, err = c.auditRepo.Create(ctx, audioInput)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// 3. Increment Paid Installments
	err = c.mpRepo.IncrementPaidInstallments(ctx, tenantID, mp.ID)
	if err != nil {
		return fmt.Errorf("failed to increment installments: %w", err)
	}

	// 4. Update Account Balance
	// Credit card: expense INCREASES balance (debt).
	newBalance := acc.BalanceCents + txInput.AmountCents
	err = c.accRepo.UpdateBalance(ctx, tenantID, acc.ID, newBalance)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	if tx != nil {
		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}
	return nil
}

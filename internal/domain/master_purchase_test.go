package domain_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestCreateMasterPurchaseInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	now := time.Now()

	tests := []struct {
		name    string
		input   domain.CreateMasterPurchaseInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: domain.CreateMasterPurchaseInput{
				FirstInstallmentDate: now,
				AccountID:            "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				CategoryID:           "01H7XRM1Z8P8P8P8P8P8P8P8P9",
				UserID:               "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Description:          "MacBook Pro",
				TotalAmountCents:     250000,
				InstallmentCount:     12,
				ClosingDay:           15,
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			input: domain.CreateMasterPurchaseInput{
				Description:      "Incomplete",
				TotalAmountCents: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid InstallmentCount too low",
			input: domain.CreateMasterPurchaseInput{
				FirstInstallmentDate: now,
				AccountID:            "ID",
				CategoryID:           "ID",
				UserID:               "ID",
				Description:          "Too low",
				TotalAmountCents:     100,
				InstallmentCount:     1, // min is 2
				ClosingDay:           10,
			},
			wantErr: true,
		},
		{
			name: "invalid InstallmentCount too high",
			input: domain.CreateMasterPurchaseInput{
				FirstInstallmentDate: now,
				AccountID:            "ID",
				CategoryID:           "ID",
				UserID:               "ID",
				Description:          "Too high",
				TotalAmountCents:     100,
				InstallmentCount:     49, // max is 48
				ClosingDay:           10,
			},
			wantErr: true,
		},
		{
			name: "invalid ClosingDay 0",
			input: domain.CreateMasterPurchaseInput{
				FirstInstallmentDate: now,
				AccountID:            "ID",
				CategoryID:           "ID",
				UserID:               "ID",
				Description:          "Invalid day",
				TotalAmountCents:     100,
				InstallmentCount:     12,
				ClosingDay:           0, // min is 1
			},
			wantErr: true,
		},
		{
			name: "invalid ClosingDay 29",
			input: domain.CreateMasterPurchaseInput{
				FirstInstallmentDate: now,
				AccountID:            "ID",
				CategoryID:           "ID",
				UserID:               "ID",
				Description:          "Invalid day",
				TotalAmountCents:     100,
				InstallmentCount:     12,
				ClosingDay:           29, // max is 28
			},
			wantErr: true,
		},
		{
			name: "invalid TotalAmountCents 0",
			input: domain.CreateMasterPurchaseInput{
				FirstInstallmentDate: now,
				AccountID:            "ID",
				CategoryID:           "ID",
				UserID:               "ID",
				Description:          "Free?",
				TotalAmountCents:     0, // gt is 0
				InstallmentCount:     12,
				ClosingDay:           15,
			},
			wantErr: true,
		},
		{
			name: "invalid TotalAmountCents negative",
			input: domain.CreateMasterPurchaseInput{
				FirstInstallmentDate: now,
				AccountID:            "ID",
				CategoryID:           "ID",
				UserID:               "ID",
				Description:          "Negative",
				TotalAmountCents:     -100,
				InstallmentCount:     12,
				ClosingDay:           15,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate.Struct(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

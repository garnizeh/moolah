package domain_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestCreateTransactionInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	now := time.Now()

	tests := []struct {
		name    string
		input   domain.CreateTransactionInput
		wantErr bool
	}{
		{
			name: "valid transaction",
			input: domain.CreateTransactionInput{
				OccurredAt:  now,
				AccountID:   "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				CategoryID:  "01H7XRM1Z8P8P8P8P8P8P8P8P9",
				UserID:      "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Description: "Weekly Groceries",
				Type:        domain.TransactionTypeExpense,
				AmountCents: 15450,
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			input: domain.CreateTransactionInput{
				Description: "Incomplete",
				AmountCents: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid transaction type",
			input: domain.CreateTransactionInput{
				OccurredAt:  now,
				AccountID:   "ID",
				CategoryID:  "ID",
				UserID:      "ID",
				Description: "Bad Type",
				Type:        "reversal",
				AmountCents: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid amount cents 0",
			input: domain.CreateTransactionInput{
				OccurredAt:  now,
				AccountID:   "ID",
				CategoryID:  "ID",
				UserID:      "ID",
				Description: "Zero",
				Type:        domain.TransactionTypeExpense,
				AmountCents: 0,
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

func TestUpdateTransactionInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	desc := "Updated"
	amount := int64(200)
	badAmount := int64(0)
	longDesc := "This description is definitely way too long for a transaction as it exceeds the two hundred and fifty-five characters limit that we have set for the description field in our database schema and validation logic for the input value objects in the domain layer."

	tests := []struct {
		input   domain.UpdateTransactionInput
		name    string
		wantErr bool
	}{
		{
			name: "valid description update",
			input: domain.UpdateTransactionInput{
				Description: &desc,
			},
			wantErr: false,
		},
		{
			name: "valid amount update",
			input: domain.UpdateTransactionInput{
				AmountCents: &amount,
			},
			wantErr: false,
		},
		{
			name: "invalid amount 0",
			input: domain.UpdateTransactionInput{
				AmountCents: &badAmount,
			},
			wantErr: true,
		},
		{
			name: "invalid description too long",
			input: domain.UpdateTransactionInput{
				Description: &longDesc,
			},
			wantErr: true,
		},
		{
			name:    "empty update (valid omitempty)",
			input:   domain.UpdateTransactionInput{},
			wantErr: false,
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

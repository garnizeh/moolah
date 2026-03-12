package domain_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestCreateAccountInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name    string
		input   domain.CreateAccountInput
		wantErr bool
	}{
		{
			name: "valid checking account",
			input: domain.CreateAccountInput{
				UserID:       "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Name:         "Main Checking",
				Type:         domain.AccountTypeChecking,
				Currency:     "USD",
				InitialCents: 10000,
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			input: domain.CreateAccountInput{
				Name: "Incomplete",
			},
			wantErr: true,
		},
		{
			name: "invalid account type",
			input: domain.CreateAccountInput{
				UserID:       "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Name:         "Invalid Type",
				Type:         "cash",
				Currency:     "USD",
				InitialCents: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid currency length",
			input: domain.CreateAccountInput{
				UserID:       "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Name:         "Invalid Currency",
				Type:         domain.AccountTypeSavings,
				Currency:     "USDOLLAR",
				InitialCents: 0,
			},
			wantErr: true,
		},
		{
			name: "name too long",
			input: domain.CreateAccountInput{
				UserID:       "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Name:         "This name is definitely longer than one hundred characters to test the validation constraint of the name field",
				Type:         domain.AccountTypeChecking,
				Currency:     "USD",
				InitialCents: 0,
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

func TestUpdateAccountInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	name := "New Name"
	currency := "EUR"
	tooLong := "This name is way too long for the account name field validation which is capped at exactly one hundred characters total"
	badCurrency := "EURO"

	tests := []struct {
		input   domain.UpdateAccountInput
		name    string
		wantErr bool
	}{
		{
			name: "valid name update",
			input: domain.UpdateAccountInput{
				Name: &name,
			},
			wantErr: false,
		},
		{
			name: "valid currency update",
			input: domain.UpdateAccountInput{
				Currency: &currency,
			},
			wantErr: false,
		},
		{
			name: "invalid name too long",
			input: domain.UpdateAccountInput{
				Name: &tooLong,
			},
			wantErr: true,
		},
		{
			name: "invalid currency length",
			input: domain.UpdateAccountInput{
				Currency: &badCurrency,
			},
			wantErr: true,
		},
		{
			name:    "empty update (valid omitempty)",
			input:   domain.UpdateAccountInput{},
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

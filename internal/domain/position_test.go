package domain_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestIncomeType_Constants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, domain.IncomeTypeNone, domain.IncomeType("none"))
	assert.Equal(t, domain.IncomeTypeDividend, domain.IncomeType("dividend"))
	assert.Equal(t, domain.IncomeTypeCoupon, domain.IncomeType("coupon"))
	assert.Equal(t, domain.IncomeTypeRent, domain.IncomeType("rent"))
	assert.Equal(t, domain.IncomeTypeInterest, domain.IncomeType("interest"))
	assert.Equal(t, domain.IncomeTypeSalary, domain.IncomeType("salary"))
}

func TestReceivableStatus_Constants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, domain.ReceivableStatusPending, domain.ReceivableStatus("pending"))
	assert.Equal(t, domain.ReceivableStatusReceived, domain.ReceivableStatus("received"))
	assert.Equal(t, domain.ReceivableStatusCancelled, domain.ReceivableStatus("cancelled"))
}

func TestCreatePositionInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	now := time.Now()

	tests := []struct {
		name    string
		input   domain.CreatePositionInput
		wantErr bool
	}{
		{
			name: "valid position",
			input: domain.CreatePositionInput{
				AssetID:        "asset-1",
				AccountID:      "account-1",
				Quantity:       decimal.NewFromFloat(10.5),
				AvgCostCents:   1000,
				LastPriceCents: 1100,
				Currency:       "USD",
				PurchasedAt:    now,
				IncomeType:     domain.IncomeTypeDividend,
			},
			wantErr: false,
		},
		{
			name: "missing asset_id",
			input: domain.CreatePositionInput{
				AccountID:      "account-1",
				Quantity:       decimal.NewFromFloat(10.5),
				AvgCostCents:   1000,
				LastPriceCents: 1100,
				Currency:       "USD",
				PurchasedAt:    now,
				IncomeType:     domain.IncomeTypeDividend,
			},
			wantErr: true,
		},
		{
			name: "invalid currency length",
			input: domain.CreatePositionInput{
				AssetID:        "asset-1",
				AccountID:      "account-1",
				Quantity:       decimal.NewFromFloat(10.5),
				AvgCostCents:   1000,
				LastPriceCents: 1100,
				Currency:       "US",
				PurchasedAt:    now,
				IncomeType:     domain.IncomeTypeDividend,
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

func TestUpdatePositionInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	qty := decimal.NewFromFloat(15.0)

	tests := []struct {
		input   domain.UpdatePositionInput
		name    string
		wantErr bool
	}{
		{
			name: "valid update",
			input: domain.UpdatePositionInput{
				Quantity: &qty,
			},
			wantErr: false,
		},
		{
			name: "invalid negative avg cost",
			input: domain.UpdatePositionInput{
				AvgCostCents: func() *int64 { v := int64(-1); return &v }(),
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

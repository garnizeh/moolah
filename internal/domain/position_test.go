package domain_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestCreatePositionInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name    string
		input   domain.CreatePositionInput
		wantErr bool
	}{
		{
			name: "valid position",
			input: domain.CreatePositionInput{
				AccountID:    "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				AssetID:      "01H7XRM1Z8P8P8P8P8P8P8P8P9",
				Quantity:     10.5,
				AvgCostCents: 15450,
			},
			wantErr: false,
		},
		{
			name: "missing account id",
			input: domain.CreatePositionInput{
				AssetID:      "01H7XRM1Z8P8P8P8P8P8P8P8P9",
				Quantity:     10.5,
				AvgCostCents: 15450,
			},
			wantErr: true,
		},
		{
			name: "missing asset id",
			input: domain.CreatePositionInput{
				AccountID:    "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				Quantity:     10.5,
				AvgCostCents: 15450,
			},
			wantErr: true,
		},
		{
			name: "invalid quantity 0",
			input: domain.CreatePositionInput{
				AccountID:    "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				AssetID:      "01H7XRM1Z8P8P8P8P8P8P8P8P9",
				Quantity:     0,
				AvgCostCents: 15450,
			},
			wantErr: true,
		},
		{
			name: "invalid quantity negative",
			input: domain.CreatePositionInput{
				AccountID:    "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				AssetID:      "01H7XRM1Z8P8P8P8P8P8P8P8P9",
				Quantity:     -1,
				AvgCostCents: 15450,
			},
			wantErr: true,
		},
		{
			name: "invalid avg cost cents negative",
			input: domain.CreatePositionInput{
				AccountID:    "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				AssetID:      "01H7XRM1Z8P8P8P8P8P8P8P8P9",
				Quantity:     10.5,
				AvgCostCents: -1,
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

	qty := 15.0
	badQty := 0.0
	cost := int64(12000)
	badCost := int64(-1)
	status := domain.PositionStatusClosed
	closedAt := time.Now()

	tests := []struct {
		input   domain.UpdatePositionInput
		name    string
		wantErr bool
	}{
		{
			name: "valid partial update",
			input: domain.UpdatePositionInput{
				Quantity: &qty,
			},
			wantErr: false,
		},
		{
			name: "valid full update",
			input: domain.UpdatePositionInput{
				Quantity:     &qty,
				AvgCostCents: &cost,
				Status:       &status,
				ClosedAt:     &closedAt,
			},
			wantErr: false,
		},
		{
			name: "invalid quantity 0",
			input: domain.UpdatePositionInput{
				Quantity: &badQty,
			},
			wantErr: true,
		},
		{
			name: "invalid avg cost negative",
			input: domain.UpdatePositionInput{
				AvgCostCents: &badCost,
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

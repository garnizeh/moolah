package domain_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestAssetType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		assetType domain.AssetType
		expected  string
	}{
		{"Stock", domain.AssetTypeStock, "stock"},
		{"Bond", domain.AssetTypeBond, "bond"},
		{"Fund", domain.AssetTypeFund, "fund"},
		{"Crypto", domain.AssetTypeCrypto, "crypto"},
		{"RealEstate", domain.AssetTypeRealEstate, "real_estate"},
		{"IncomeSource", domain.AssetTypeIncomeSource, "income_source"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, string(tc.assetType))
		})
	}
}

func TestCreateAssetInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	strPtr := func(s string) *string { return &s }

	type testCase struct {
		input   domain.CreateAssetInput
		name    string
		wantErr bool
	}

	tests := []testCase{
		{
			name: "Valid Input",
			input: domain.CreateAssetInput{
				Ticker:    "AAPL",
				Name:      "Apple Inc.",
				AssetType: domain.AssetTypeStock,
				Currency:  "USD",
			},
			wantErr: false,
		},
		{
			name: "Missing Ticker",
			input: domain.CreateAssetInput{
				Name:      "Apple Inc.",
				AssetType: domain.AssetTypeStock,
				Currency:  "USD",
			},
			wantErr: true,
		},
		{
			name: "Ticker Too Long",
			input: domain.CreateAssetInput{
				Ticker:    "VERYLONGTICKERWITHTOOMANYCHARS",
				Name:      "Apple Inc.",
				AssetType: domain.AssetTypeStock,
				Currency:  "USD",
			},
			wantErr: true,
		},
		{
			name: "Missing Name",
			input: domain.CreateAssetInput{
				Ticker:    "AAPL",
				AssetType: domain.AssetTypeStock,
				Currency:  "USD",
			},
			wantErr: true,
		},
		{
			name: "Missing AssetType",
			input: domain.CreateAssetInput{
				Ticker:   "AAPL",
				Name:     "Apple Inc.",
				Currency: "USD",
			},
			wantErr: true,
		},
		{
			name: "Invalid Currency Length",
			input: domain.CreateAssetInput{
				Ticker:    "AAPL",
				Name:      "Apple Inc.",
				AssetType: domain.AssetTypeStock,
				Currency:  "USDOLLAR",
			},
			wantErr: true,
		},
		{
			name: "Valid ISIN",
			input: domain.CreateAssetInput{
				Ticker:    "AAPL",
				Name:      "Apple Inc.",
				AssetType: domain.AssetTypeStock,
				Currency:  "USD",
				ISIN:      strPtr("US0378331005"),
			},
			wantErr: false,
		},
		{
			name: "ISIN Too Long",
			input: domain.CreateAssetInput{
				Ticker:    "AAPL",
				Name:      "Apple Inc.",
				AssetType: domain.AssetTypeStock,
				Currency:  "USD",
				ISIN:      strPtr("INVALIDISINTOOLONG"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate.Struct(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpsertTenantAssetConfigInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	strPtr := func(s string) *string { return &s }

	type testCase struct {
		input   domain.UpsertTenantAssetConfigInput
		name    string
		wantErr bool
	}

	tests := []testCase{
		{
			name: "Valid Input",
			input: domain.UpsertTenantAssetConfigInput{
				AssetID:  "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				Currency: strPtr("BRL"),
			},
			wantErr: false,
		},
		{
			name: "Missing AssetID",
			input: domain.UpsertTenantAssetConfigInput{
				Currency: strPtr("BRL"),
			},
			wantErr: true,
		},
		{
			name: "Invalid Overridden Currency Length",
			input: domain.UpsertTenantAssetConfigInput{
				AssetID:  "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				Currency: strPtr("BRLL"),
			},
			wantErr: true,
		},
		{
			name: "Valid Optional Name",
			input: domain.UpsertTenantAssetConfigInput{
				AssetID: "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				Name:    strPtr("Personalized Name"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate.Struct(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

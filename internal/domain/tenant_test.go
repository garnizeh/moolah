package domain_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestCreateTenantInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name    string
		input   domain.CreateTenantInput
		wantErr bool
	}{
		{
			name: "valid tenant",
			input: domain.CreateTenantInput{
				Name: "The Simpson Household",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			input: domain.CreateTenantInput{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "name too short",
			input: domain.CreateTenantInput{
				Name: "A",
			},
			wantErr: true,
		},
		{
			name: "name too long",
			input: domain.CreateTenantInput{
				Name: "This is a very long name that exceeds the maximum limit of one hundred characters allowed for a tenant name in the system",
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

func TestUpdateTenantInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	name := "Updated Household"
	plan := domain.TenantPlanPremium
	tooShort := "A"

	tests := []struct {
		input   domain.UpdateTenantInput
		name    string
		wantErr bool
	}{
		{
			name: "valid name update",
			input: domain.UpdateTenantInput{
				Name: &name,
			},
			wantErr: false,
		},
		{
			name: "valid plan update",
			input: domain.UpdateTenantInput{
				Plan: &plan,
			},
			wantErr: false,
		},
		{
			name: "invalid name too short",
			input: domain.UpdateTenantInput{
				Name: &tooShort,
			},
			wantErr: true,
		},
		{
			name:    "empty update (valid omitempty)",
			input:   domain.UpdateTenantInput{},
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

package domain_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestCreateUserInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		input   domain.CreateUserInput
		name    string
		wantErr bool
	}{
		{
			name: "valid user input",
			input: domain.CreateUserInput{
				TenantID: "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				Email:    "test@example.com",
				Name:     "John Doe",
				Role:     domain.RoleMember,
			},
			wantErr: false,
		},
		{
			name: "missing email",
			input: domain.CreateUserInput{
				TenantID: "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				Name:     "John Doe",
				Role:     domain.RoleMember,
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			input: domain.CreateUserInput{
				TenantID: "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				Email:    "not-an-email",
				Name:     "John Doe",
				Role:     domain.RoleMember,
			},
			wantErr: true,
		},
		{
			name: "name too short",
			input: domain.CreateUserInput{
				TenantID: "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				Email:    "test@example.com",
				Name:     "A",
				Role:     domain.RoleMember,
			},
			wantErr: true,
		},
		{
			name: "missing role",
			input: domain.CreateUserInput{
				TenantID: "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				Email:    "test@example.com",
				Name:     "John Doe",
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

func TestUpdateUserInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	name := "Jane Doe"
	role := domain.RoleAdmin
	tooShort := "A"

	tests := []struct {
		input   domain.UpdateUserInput
		name    string
		wantErr bool
	}{
		{
			name: "valid name update",
			input: domain.UpdateUserInput{
				Name: &name,
			},
			wantErr: false,
		},
		{
			name: "valid role update",
			input: domain.UpdateUserInput{
				Role: &role,
			},
			wantErr: false,
		},
		{
			name: "invalid name too short",
			input: domain.UpdateUserInput{
				Name: &tooShort,
			},
			wantErr: true,
		},
		{
			name:    "empty update (valid omitempty)",
			input:   domain.UpdateUserInput{},
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

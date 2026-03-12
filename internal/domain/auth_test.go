package domain_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestCreateOTPRequestInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	now := time.Now()

	tests := []struct {
		name    string
		input   domain.CreateOTPRequestInput
		wantErr bool
	}{
		{
			name: "valid otp request",
			input: domain.CreateOTPRequestInput{
				ExpiresAt: now.Add(time.Minute * 10),
				Email:     "user@example.com",
				CodeHash:  "$2a$10$vI8p...",
			},
			wantErr: false,
		},
		{
			name: "missing expires_at",
			input: domain.CreateOTPRequestInput{
				Email:    "user@example.com",
				CodeHash: "hash",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			input: domain.CreateOTPRequestInput{
				ExpiresAt: now,
				Email:     "not-an-email",
				CodeHash:  "hash",
			},
			wantErr: true,
		},
		{
			name: "missing code hash",
			input: domain.CreateOTPRequestInput{
				ExpiresAt: now,
				Email:     "user@example.com",
				CodeHash:  "",
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

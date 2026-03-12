package domain_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/garnizeh/moolah/internal/domain"
)

func TestCreateAuditLogInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name    string
		input   domain.CreateAuditLogInput
		wantErr bool
	}{
		{
			name: "valid full input",
			input: domain.CreateAuditLogInput{
				TenantID:   "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				ActorID:    "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Action:     domain.AuditActionCreate,
				EntityType: "transaction",
				EntityID:   "01H7XRM1Z8P8P8P8P8P8P8P8P1",
				IPAddress:  "127.0.0.1",
				UserAgent:  "Mozilla/5.0",
				ActorRole:  domain.RoleMember,
			},
			wantErr: false,
		},
		{
			name: "valid minimal input",
			input: domain.CreateAuditLogInput{
				TenantID:   "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				ActorID:    "SYSTEM",
				Action:     domain.AuditActionOTPRequested,
				EntityType: "auth",
				ActorRole:  domain.RoleSysadmin,
			},
			wantErr: false,
		},
		{
			name: "missing tenant_id",
			input: domain.CreateAuditLogInput{
				ActorID:    "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Action:     domain.AuditActionCreate,
				EntityType: "account",
				ActorRole:  domain.RoleAdmin,
			},
			wantErr: true,
		},
		{
			name: "missing actor_id",
			input: domain.CreateAuditLogInput{
				TenantID:   "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				Action:     domain.AuditActionUpdate,
				EntityType: "category",
				ActorRole:  domain.RoleMember,
			},
			wantErr: true,
		},
		{
			name: "missing action",
			input: domain.CreateAuditLogInput{
				TenantID:   "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				ActorID:    "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				EntityType: "transaction",
				ActorRole:  domain.RoleMember,
			},
			wantErr: true,
		},
		{
			name: "missing entity_type",
			input: domain.CreateAuditLogInput{
				TenantID:  "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				ActorID:   "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Action:    domain.AuditActionSoftDelete,
				ActorRole: domain.RoleAdmin,
			},
			wantErr: true,
		},
		{
			name: "missing actor_role",
			input: domain.CreateAuditLogInput{
				TenantID:   "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				ActorID:    "01H7XRM1Z8P8P8P8P8P8P8P8P0",
				Action:     domain.AuditActionLogin,
				EntityType: "user",
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

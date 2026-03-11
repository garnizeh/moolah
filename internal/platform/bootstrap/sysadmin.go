package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/pkg/config"
	"github.com/garnizeh/moolah/pkg/ulid"
)

// EnsureSysadmin creates the system tenant and sysadmin user if they do not exist.
// It is idempotent: calling it multiple times has no side effects after the first run.
func EnsureSysadmin(ctx context.Context, q sqlc.Querier, cfg *config.Config) error {
	// 1. Check if sysadmin with this email already exists
	existing, err := q.GetUserByEmail(ctx, cfg.SysadminEmail)
	if err == nil {
		if existing.Role == sqlc.UserRoleSysadmin {
			slog.Info("sysadmin already exists, skipping bootstrap", "email", cfg.SysadminEmail)
			return nil
		}
		// If user exists but is not sysadmin, we should probably warn or error,
		// but for now let's just err out as it's a conflict in expectations.
		return fmt.Errorf("bootstrap: user %s exists but is not a sysadmin (role: %s)", cfg.SysadminEmail, existing.Role)
	}

	if !errors.Is(repository.TranslateError(err), domain.ErrNotFound) {
		return fmt.Errorf("bootstrap: failed to check existing sysadmin: %w", err)
	}

	slog.Info("bootstrapping sysadmin...", "email", cfg.SysadminEmail, "tenant", cfg.SysadminTenantName)

	// 2. Create system tenant
	tenantID := ulid.New()
	_, err = q.CreateTenant(ctx, sqlc.CreateTenantParams{
		ID:   tenantID,
		Name: cfg.SysadminTenantName,
		Plan: sqlc.TenantPlanFree,
	})
	if err != nil {
		return fmt.Errorf("bootstrap: failed to create system tenant: %w", err)
	}

	// 3. Create sysadmin user
	_, err = q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       ulid.New(),
		TenantID: tenantID,
		Email:    cfg.SysadminEmail,
		Name:     "Sysadmin",
		Role:     sqlc.UserRoleSysadmin,
	})
	if err != nil {
		return fmt.Errorf("bootstrap: failed to create sysadmin user: %w", err)
	}

	slog.Info("sysadmin bootstrapped successfully", "email", cfg.SysadminEmail, "tenant", cfg.SysadminTenantName)
	return nil
}

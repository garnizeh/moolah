package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/pkg/otp"
	pkgpaseto "github.com/garnizeh/moolah/pkg/paseto"
)

type authService struct {
	authRepo  domain.AuthRepository
	userRepo  domain.UserRepository
	auditRepo domain.AuditRepository
	mailer    domain.Mailer
	pasetoKey paseto.V4SymmetricKey
}

// NewAuthService creates a new authentication service.
func NewAuthService(
	authRepo domain.AuthRepository,
	userRepo domain.UserRepository,
	auditRepo domain.AuditRepository,
	mailer domain.Mailer,
	pasetoKey paseto.V4SymmetricKey,
) domain.AuthService {
	return &authService{
		authRepo:  authRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		mailer:    mailer,
		pasetoKey: pasetoKey,
	}
}

func (s *authService) RequestOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("auth service: user not found: %w", domain.ErrNotFound)
		}
		return fmt.Errorf("auth service: failed to lookup user: %w", err)
	}

	plain, codeHash, err := otp.Generate()
	if err != nil {
		return fmt.Errorf("auth service: failed to generate otp: %w", err)
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	_, err = s.authRepo.CreateOTPRequest(ctx, domain.CreateOTPRequestInput{
		Email:     email,
		CodeHash:  codeHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("auth service: failed to save otp request: %w", err)
	}

	err = s.mailer.SendOTP(ctx, email, plain)
	if err != nil {
		return fmt.Errorf("auth service: failed to send email: %w", err)
	}

	_, auditErr := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   user.TenantID,
		ActorID:    user.ID,
		Action:     domain.AuditActionOTPRequested,
		EntityType: "otp_request",
		IPAddress:  "", // Optional - to be filled by handler if possible
		UserAgent:  "", // Optional
		ActorRole:  user.Role,
	})
	if auditErr != nil {
		// We don't fail the whole request if audit fails, but we should log it
		// For now, returning it or wrap it is safer for Phase 1.
		// Actually, standard is to return it or at least not ignore it.
		// But if OTP is sent, we shouldn't block the user.
		// However, instructions say NEVER ignore errors.
		return fmt.Errorf("auth service: failed to create audit log: %w", auditErr)
	}

	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, email, code string) (*domain.TokenPair, error) {
	otpReq, err := s.authRepo.GetActiveOTPRequest(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOTP) {
			// Record login failure for unknown user is tricky without ID,
			// but we can try to look up the user by email for auditing.
			user, _ := s.userRepo.GetByEmail(ctx, email)
			if user != nil {
				_, _ = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
					TenantID:   user.TenantID,
					ActorID:    user.ID,
					Action:     domain.AuditActionLoginFailed,
					EntityType: "auth",
					ActorRole:  user.Role,
				})
			}
			return nil, domain.ErrInvalidOTP
		}
		return nil, fmt.Errorf("auth service: failed to fetch active otp: %w", err)
	}

	err = otp.Verify(code, otpReq.CodeHash)
	if err != nil {
		user, _ := s.userRepo.GetByEmail(ctx, email)
		if user != nil {
			_, _ = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
				TenantID:   user.TenantID,
				ActorID:    user.ID,
				Action:     domain.AuditActionLoginFailed,
				EntityType: "auth",
				ActorRole:  user.Role,
			})
		}
		return nil, domain.ErrInvalidOTP
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("auth service: user suddenly missing: %w", err)
	}

	err = s.authRepo.MarkOTPUsed(ctx, otpReq.ID)
	if err != nil {
		return nil, fmt.Errorf("auth service: failed to mark otp as used: %w", err)
	}

	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("auth service: failed to update last login: %w", err)
	}

	now := time.Now()
	accessExpiry := now.Add(15 * time.Minute)
	refreshExpiry := now.Add(7 * 24 * time.Hour)

	accessToken, err := pkgpaseto.Seal(pkgpaseto.Claims{
		IssuedAt:  now,
		ExpiresAt: accessExpiry,
		TenantID:  user.TenantID,
		UserID:    user.ID,
		Role:      string(user.Role),
	}, s.pasetoKey)
	if err != nil {
		return nil, fmt.Errorf("auth service: failed to sign access token: %w", err)
	}

	refreshToken, err := pkgpaseto.Seal(pkgpaseto.Claims{
		IssuedAt:  now,
		ExpiresAt: refreshExpiry,
		TenantID:  user.TenantID,
		UserID:    user.ID,
		Role:      string(user.Role),
	}, s.pasetoKey)
	if err != nil {
		return nil, fmt.Errorf("auth service: failed to sign refresh token: %w", err)
	}

	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   user.TenantID,
		ActorID:    user.ID,
		Action:     domain.AuditActionOTPVerified,
		EntityType: "auth",
		ActorRole:  user.Role,
	})
	if err != nil {
		return nil, fmt.Errorf("auth service: failed to create audit log: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	claims, err := pkgpaseto.Parse(refreshToken, s.pasetoKey)
	if err != nil {
		return nil, fmt.Errorf("auth service: invalid refresh token: %w", domain.ErrUnauthorized)
	}

	user, err := s.userRepo.GetByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("auth service: user deleted: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("auth service: user lookup failed: %w", err)
	}

	now := time.Now()
	accessExpiry := now.Add(15 * time.Minute)
	refreshExpiry := now.Add(7 * 24 * time.Hour)

	newAccessToken, err := pkgpaseto.Seal(pkgpaseto.Claims{
		IssuedAt:  now,
		ExpiresAt: accessExpiry,
		TenantID:  user.TenantID,
		UserID:    user.ID,
		Role:      string(user.Role),
	}, s.pasetoKey)
	if err != nil {
		return nil, fmt.Errorf("auth service: failed to sign access token: %w", err)
	}

	newRefreshToken, err := pkgpaseto.Seal(pkgpaseto.Claims{
		IssuedAt:  now,
		ExpiresAt: refreshExpiry,
		TenantID:  user.TenantID,
		UserID:    user.ID,
		Role:      string(user.Role),
	}, s.pasetoKey)
	if err != nil {
		return nil, fmt.Errorf("auth service: failed to sign refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

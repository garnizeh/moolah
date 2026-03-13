package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	pkgpaseto "github.com/garnizeh/moolah/pkg/paseto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_RequestOTP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "test@example.com"
	user := &domain.User{
		ID:       "user_123",
		TenantID: "tenant_456",
		Email:    email,
		Role:     domain.RoleMember,
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)
		mailer := new(mocks.Mailer)
		key := paseto.NewV4SymmetricKey()

		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		authRepo.On("CreateOTPRequest", ctx, mock.MatchedBy(func(input domain.CreateOTPRequestInput) bool {
			return input.Email == email && input.CodeHash != ""
		})).Return(&domain.OTPRequest{ID: "otp_1"}, nil)
		mailer.On("SendOTP", ctx, email, mock.MatchedBy(func(s string) bool { return len(s) == 6 })).Return(nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(input domain.CreateAuditLogInput) bool {
			return input.ActorID == user.ID && input.Action == domain.AuditActionOTPRequested
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, mailer, key)
		err := svc.RequestOTP(ctx, email)

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
		authRepo.AssertExpectations(t)
		mailer.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mocks.UserRepository)
		userRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrNotFound)

		svc := service.NewAuthService(nil, userRepo, nil, nil, paseto.NewV4SymmetricKey())
		err := svc.RequestOTP(ctx, email)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		userRepo.AssertExpectations(t)
	})

	t.Run("UserLookupError", func(t *testing.T) {
		t.Parallel()
		userRepo := new(mocks.UserRepository)
		userRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("db error"))

		svc := service.NewAuthService(nil, userRepo, nil, nil, paseto.NewV4SymmetricKey())
		err := svc.RequestOTP(ctx, email)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to lookup user")
	})

	t.Run("CreateOTPError", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		key := paseto.NewV4SymmetricKey()

		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		authRepo.On("CreateOTPRequest", ctx, mock.Anything).Return(nil, errors.New("db error"))

		svc := service.NewAuthService(authRepo, userRepo, nil, nil, key)
		err := svc.RequestOTP(ctx, email)
		require.Error(t, err)
	})

	t.Run("MailerError", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		mailer := new(mocks.Mailer)
		key := paseto.NewV4SymmetricKey()

		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		authRepo.On("CreateOTPRequest", ctx, mock.Anything).Return(&domain.OTPRequest{ID: "otp_1"}, nil)
		mailer.On("SendOTP", ctx, email, mock.Anything).Return(errors.New("smtp error"))

		svc := service.NewAuthService(authRepo, userRepo, nil, mailer, key)
		err := svc.RequestOTP(ctx, email)
		require.Error(t, err)
	})

	t.Run("AuditError", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)
		mailer := new(mocks.Mailer)
		key := paseto.NewV4SymmetricKey()

		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		authRepo.On("CreateOTPRequest", ctx, mock.Anything).Return(&domain.OTPRequest{ID: "otp_1"}, nil)
		mailer.On("SendOTP", ctx, email, mock.Anything).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit error"))

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, mailer, key)
		err := svc.RequestOTP(ctx, email)
		// We expect NoError because the service should not fail the whole request
		// if auditing fails (it should only logged it for resilience).
		require.NoError(t, err)
	})
}

func TestAuthService_VerifyOTP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "test@example.com"
	plainCode := "123456"
	codeHashBytes, err := bcrypt.GenerateFromPassword([]byte(plainCode), 4)
	require.NoError(t, err)
	codeHash := string(codeHashBytes)
	user := &domain.User{
		ID:       "user_123",
		TenantID: "tenant_456",
		Email:    email,
		Role:     domain.RoleMember,
	}
	otpReq := &domain.OTPRequest{
		ID:       "otp_1",
		Email:    email,
		CodeHash: codeHash,
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)
		key := paseto.NewV4SymmetricKey()

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(otpReq, nil)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		authRepo.On("MarkOTPUsed", ctx, otpReq.ID).Return(nil)
		userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(input domain.CreateAuditLogInput) bool {
			return input.ActorID == user.ID && input.Action == domain.AuditActionOTPVerified
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, nil, key)
		tokens, err := svc.VerifyOTP(ctx, email, plainCode)

		require.NoError(t, err)
		require.NotNil(t, tokens)
		require.NotEmpty(t, tokens.AccessToken)
		require.NotEmpty(t, tokens.RefreshToken)

		authRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("OTPNotFound_WithAudit", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(nil, domain.ErrInvalidOTP)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(input domain.CreateAuditLogInput) bool {
			return input.Action == domain.AuditActionLoginFailed
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, "111111")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidOTP)
	})

	t.Run("OTPNotFound_UserMissingForAudit", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(nil, domain.ErrInvalidOTP)
		userRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrNotFound)

		svc := service.NewAuthService(authRepo, userRepo, nil, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, "111111")
		require.Error(t, err)
	})

	t.Run("FetchOTPError", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		authRepo.On("GetActiveOTPRequest", ctx, email).Return(nil, errors.New("db error"))

		svc := service.NewAuthService(authRepo, nil, nil, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, plainCode)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to fetch active otp")
	})

	t.Run("InvalidOTPCode_Audited", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(otpReq, nil)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(input domain.CreateAuditLogInput) bool {
			return input.ActorID == user.ID && input.Action == domain.AuditActionLoginFailed
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, "wrong")

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidOTP)
	})

	t.Run("InvalidOTPCode_UserMissing", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(otpReq, nil)
		userRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrNotFound)

		svc := service.NewAuthService(authRepo, userRepo, nil, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, "wrong")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidOTP)
	})

	t.Run("MarkOTPUsedError", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(otpReq, nil)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		authRepo.On("MarkOTPUsed", ctx, otpReq.ID).Return(errors.New("db error"))

		svc := service.NewAuthService(authRepo, userRepo, nil, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, plainCode)
		require.Error(t, err)
	})

	t.Run("UpdateLastLoginError", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(otpReq, nil)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		authRepo.On("MarkOTPUsed", ctx, otpReq.ID).Return(nil)
		userRepo.On("UpdateLastLogin", ctx, user.ID).Return(errors.New("db error"))

		svc := service.NewAuthService(authRepo, userRepo, nil, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, plainCode)
		require.Error(t, err)
	})

	t.Run("AuditError_Verify", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)
		key := paseto.NewV4SymmetricKey()

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(otpReq, nil)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		authRepo.On("MarkOTPUsed", ctx, otpReq.ID).Return(nil)
		userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit error"))

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, nil, key)
		_, err := svc.VerifyOTP(ctx, email, plainCode)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create audit log")
	})

	t.Run("OTPNotFound_AuditLogFailure", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(nil, domain.ErrInvalidOTP)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit log failure"))

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, "111111")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidOTP)
	})

	t.Run("InvalidOTPCode_AuditLogFailure", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(otpReq, nil)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit log failure"))

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, "wrong")

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidOTP)
	})

	t.Run("UserSuddenlyMissing", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(otpReq, nil)
		userRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("suddenly missing"))

		svc := service.NewAuthService(authRepo, userRepo, nil, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, plainCode)
		require.Error(t, err)
		require.Contains(t, err.Error(), "user suddenly missing")
	})

	t.Run("AuditCreationError_FailedLogin", func(t *testing.T) {
		t.Parallel()
		authRepo := new(mocks.AuthRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		authRepo.On("GetActiveOTPRequest", ctx, email).Return(nil, domain.ErrInvalidOTP)
		userRepo.On("GetByEmail", ctx, email).Return(user, nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(input domain.CreateAuditLogInput) bool {
			return input.Action == domain.AuditActionLoginFailed
		})).Return(nil, errors.New("audit creation failed"))

		svc := service.NewAuthService(authRepo, userRepo, auditRepo, nil, paseto.NewV4SymmetricKey())
		_, err := svc.VerifyOTP(ctx, email, "111111")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid or expired OTP")
		// Since RequestOTP logic for LoginFailed audit failure just returns the original error
	})
}


func TestAuthService_RefreshToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := paseto.NewV4SymmetricKey()
	user := &domain.User{
		ID:       "user_123",
		TenantID: "tenant_456",
		Role:     domain.RoleMember,
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mocks.UserRepository)
		svc := service.NewAuthService(nil, userRepo, nil, nil, key)

		now := time.Now()
		refreshToken, err := pkgpaseto.Seal(pkgpaseto.Claims{
			IssuedAt:  now,
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			TenantID:  user.TenantID,
			UserID:    user.ID,
			Role:      string(user.Role),
		}, key)
		require.NoError(t, err)

		userRepo.On("GetByID", ctx, user.TenantID, user.ID).Return(user, nil)

		tokens, err := svc.RefreshToken(ctx, refreshToken)

		require.NoError(t, err)
		require.NotNil(t, tokens)
		require.NotEmpty(t, tokens.AccessToken)
		require.NotEmpty(t, tokens.RefreshToken)
		userRepo.AssertExpectations(t)
	})

	t.Run("UserLookupError", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mocks.UserRepository)
		svc := service.NewAuthService(nil, userRepo, nil, nil, key)

		now := time.Now()
		refreshToken, err := pkgpaseto.Seal(pkgpaseto.Claims{
			IssuedAt:  now,
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			TenantID:  user.TenantID,
			UserID:    user.ID,
			Role:      string(user.Role),
		}, key)
		require.NoError(t, err)

		userRepo.On("GetByID", ctx, user.TenantID, user.ID).Return(nil, errors.New("db error"))

		_, err = svc.RefreshToken(ctx, refreshToken)

		require.Error(t, err)
		require.Contains(t, err.Error(), "user lookup failed")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		t.Parallel()

		svc := service.NewAuthService(nil, nil, nil, nil, key)
		_, err := svc.RefreshToken(ctx, "v4.local.invalid-token")

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("UserDeleted", func(t *testing.T) {
		t.Parallel()
		userRepo := new(mocks.UserRepository)
		svc := service.NewAuthService(nil, userRepo, nil, nil, key)

		now := time.Now()
		refreshToken, err := pkgpaseto.Seal(pkgpaseto.Claims{
			IssuedAt:  now,
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			TenantID:  user.TenantID,
			UserID:    user.ID,
			Role:      string(user.Role),
		}, key)
		require.NoError(t, err)

		userRepo.On("GetByID", ctx, user.TenantID, user.ID).Return(nil, domain.ErrNotFound)

		_, err = svc.RefreshToken(ctx, refreshToken)
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("UserLookupError", func(t *testing.T) {
		t.Parallel()
		userRepo := new(mocks.UserRepository)
		svc := service.NewAuthService(nil, userRepo, nil, nil, key)

		now := time.Now()
		refreshToken, err := pkgpaseto.Seal(pkgpaseto.Claims{
			IssuedAt:  now,
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			TenantID:  user.TenantID,
			UserID:    user.ID,
			Role:      string(user.Role),
		}, key)
		require.NoError(t, err)

		userRepo.On("GetByID", ctx, user.TenantID, user.ID).Return(nil, errors.New("db error"))

		_, err = svc.RefreshToken(ctx, refreshToken)
		require.Error(t, err)
		require.Contains(t, err.Error(), "user lookup failed")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		t.Parallel()

		svc := service.NewAuthService(nil, nil, nil, nil, key)
		_, err := svc.RefreshToken(ctx, "v4.local.invalid-token")

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}

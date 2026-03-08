package mocks

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// AuthService is a testify/mock implementation of domain.AuthService.
type AuthService struct {
	mock.Mock
}

func (m *AuthService) RequestOTP(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AuthService.RequestOTP: %w", e)
	}
	return nil
}

func (m *AuthService) VerifyOTP(ctx context.Context, email, code string) (*domain.TokenPair, error) {
	args := m.Called(ctx, email, code)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AuthService.VerifyOTP: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.TokenPair)
	if !ok {
		return nil, fmt.Errorf("mock AuthService.VerifyOTP: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AuthService.RefreshToken: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.TokenPair)
	if !ok {
		return nil, fmt.Errorf("mock AuthService.RefreshToken: unexpected type %T", args.Get(0))
	}
	return res, err
}

var _ domain.AuthService = (*AuthService)(nil)

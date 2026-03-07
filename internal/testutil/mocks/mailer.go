package mocks

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/mock"

	"github.com/garnizeh/moolah/internal/domain"
)

// Mailer is a testify/mock implementation of domain.Mailer.
type Mailer struct {
	mock.Mock
}

func (m *Mailer) SendOTP(ctx context.Context, email, code string) error {
	args := m.Called(ctx, email, code)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock mailer SendOTP: %w", err)
	}
	return nil
}

var _ domain.Mailer = (*Mailer)(nil)

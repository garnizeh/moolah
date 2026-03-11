//go:build integration

package mailer

import (
	"context"
	"sync"
)

// CapturingMailer is a domain.Mailer implementation that captures sent OTP codes.
// Use in integration/smoke tests to inspect OTP codes without a real SMTP server.
type CapturingMailer struct {
	otps map[string]string // email → most recent OTP code; pointer field first for GC efficiency
	mu   sync.Mutex
}

// NewCapturingMailer constructs a CapturingMailer.
func NewCapturingMailer() *CapturingMailer {
	return &CapturingMailer{
		otps: make(map[string]string),
	}
}

// SendOTP captures the OTP code for later retrieval in tests.
func (m *CapturingMailer) SendOTP(_ context.Context, email, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.otps[email] = code
	return nil
}

// OTPFor returns the most recently sent OTP code for the given email.
// Returns an empty string if no OTP has been sent.
func (m *CapturingMailer) OTPFor(email string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.otps[email]
}

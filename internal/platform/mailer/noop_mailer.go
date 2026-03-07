package mailer

import "context"

// NoopMailer is a domain.Mailer implementation that discards all emails.
// Use in unit tests and local development.
type NoopMailer struct{}

// NewNoopMailer constructs a NoopMailer.
func NewNoopMailer() *NoopMailer {
	return &NoopMailer{}
}

// SendOTP is a no-op implementation of domain.Mailer.
func (n *NoopMailer) SendOTP(_ context.Context, _, _ string) error {
	return nil
}

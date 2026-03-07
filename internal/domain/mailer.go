package domain

import "context"

// Mailer defines the contract for sending application emails.
// Implementations: SMTPMailer (production), NoopMailer (tests).
type Mailer interface {
	// SendOTP sends a one-time password code to the given email address.
	SendOTP(ctx context.Context, to, code string) error
}

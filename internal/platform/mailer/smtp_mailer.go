package mailer

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/garnizeh/moolah/internal/domain"
)

// SMTPMailer implements domain.Mailer via net/smtp.
type SMTPMailer struct {
	host     string
	username string
	password string
	from     string
	port     int
}

// NewSMTPMailer constructs an SMTPMailer. Returns an error if required fields are empty.
func NewSMTPMailer(host string, port int, username, password, from string) (*SMTPMailer, error) {
	if host == "" {
		return nil, fmt.Errorf("mailer: host is required")
	}
	if from == "" {
		return nil, fmt.Errorf("mailer: from address is required")
	}
	// Note: username/password can be empty for local testing (e.g., Mailhog)

	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}, nil
}

// SendOTP sends the OTP code to the recipient via SMTP.
func (m *SMTPMailer) SendOTP(ctx context.Context, to, code string) error {
	if to == "" {
		return fmt.Errorf("mailer: recipient address is required")
	}
	if code == "" {
		return fmt.Errorf("mailer: OTP code is required")
	}

	subject := "Your Moolah verification code"
	body := fmt.Sprintf("Your one-time login code is: %s\n\nThis code expires in 10 minutes. Do not share it with anyone.", code)

	msg := fmt.Appendf(nil, "To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\r\n"+
		"\r\n"+
		"%s\r\n", to, m.from, subject, body)

	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	// Auth is optional for some SMTP servers (like Mailhog)
	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- smtp.SendMail(addr, auth, m.from, []string{to}, msg)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("mailer: context cancelled: %w", ctx.Err())
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("mailer: failed to send email: %w", err)
		}
	}

	return nil
}

// Ensure interface compliance at compile time.
var _ domain.Mailer = (*SMTPMailer)(nil)

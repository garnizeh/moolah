package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Helper to clear env vars after test
	setFullEnv := func(t *testing.T) {
		t.Helper()
		t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
		t.Setenv("REDIS_ADDR", "localhost:6379")
		t.Setenv("PASETO_SECRET_KEY", "70617365746f2d7365637265742d6b65792d6e6f742d6c65616b6564")
		t.Setenv("SMTP_HOST", "smtp.mailtrap.io")
		t.Setenv("SMTP_USER", "user")
		t.Setenv("SMTP_PASSWORD", "pass")
		t.Setenv("EMAIL_FROM", "no-reply@moolah.io")
		t.Setenv("SYSADMIN_EMAIL", "admin@moolah.io")
	}

	t.Run("success with all required and default optional", func(t *testing.T) {
		setFullEnv(t)

		cfg := Load()

		assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.DatabaseURL)
		assert.Equal(t, "8080", cfg.HTTPPort)
		assert.Equal(t, 10*time.Second, cfg.ReadTimeout)
		assert.Equal(t, 587, cfg.SMTPPort)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("success with custom optional", func(t *testing.T) {
		setFullEnv(t)
		t.Setenv("HTTP_PORT", "9000")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("SMTP_PORT", "2525")
		t.Setenv("TOKEN_TTL", "1h")

		cfg := Load()

		assert.Equal(t, "9000", cfg.HTTPPort)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, 2525, cfg.SMTPPort)
		assert.Equal(t, time.Hour, cfg.TokenTTL)
	})

	requiredVars := []string{
		"DATABASE_URL",
		"REDIS_ADDR",
		"PASETO_SECRET_KEY",
		"SMTP_HOST",
		"SMTP_USER",
		"SMTP_PASSWORD",
		"EMAIL_FROM",
		"SYSADMIN_EMAIL",
	}

	for _, v := range requiredVars {
		t.Run("panic if "+v+" is missing", func(t *testing.T) {
			setFullEnv(t)
			t.Setenv(v, "") // Clear the required var

			assert.PanicsWithValue(t, "config: "+v+" is required", func() {
				Load()
			})
		})
	}

	t.Run("panic on invalid duration", func(t *testing.T) {
		setFullEnv(t)
		t.Setenv("READ_TIMEOUT", "invalid")

		assert.Panics(t, func() {
			Load()
		})
	})

	t.Run("panic on invalid port", func(t *testing.T) {
		setFullEnv(t)
		t.Setenv("SMTP_PORT", "not-a-number")

		assert.Panics(t, func() {
			Load()
		})
	})
}

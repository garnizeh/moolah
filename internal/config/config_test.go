package config

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/garnizeh/moolah/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestConfig_Log_JSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	// create a test logger and set as default
	l := logger.New(&buf, "info", "json")
	prev := slog.Default()
	slog.SetDefault(l)
	defer slog.SetDefault(prev)

	cfg := &Config{
		DatabaseURL:        "postgres://user:pass@localhost:5432/db",
		RedisAddr:          "localhost:6379",
		RedisPassword:      "rpass",
		PasetoSecretKey:    "deadbeef",
		SMTPHost:           "mail.local",
		SMTPUser:           "smtp-user",
		SMTPPassword:       "smtp-pass",
		EmailFrom:          "noreply@example.com",
		HTTPPort:           "8080",
		LogLevel:           "info",
		LogFormat:          "json",
		SysadminEmail:      "admin@example.com",
		SysadminTenantName: "System",
		ReadTimeout:        1 * time.Second,
		WriteTimeout:       2 * time.Second,
		ShutdownTimeout:    3 * time.Second,
		TokenTTL:           4 * time.Second,
		SMTPPort:           587,
	}

	cfg.Log(t.Context())

	out := buf.Bytes()
	require.NotEmpty(t, out)

	var data map[string]any
	err := json.Unmarshal(out, &data)
	require.NoError(t, err)

	// check a few representative fields were logged
	assert.Equal(t, cfg.DatabaseURL, data["DatabaseURL"])
	assert.Equal(t, cfg.RedisAddr, data["RedisAddr"])
	assert.Equal(t, cfg.EmailFrom, data["EmailFrom"])
	assert.Equal(t, cfg.HTTPPort, data["HTTPPort"])
	assert.Equal(t, float64(cfg.SMTPPort), data["SMTPPort"])
}

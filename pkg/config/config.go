// Package config provides functionality to load and manage application configuration from environment variables.
package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds all the configuration values for the application.
type Config struct {
	// Database
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	PasetoSecretKey    string
	SMTPHost           string
	SMTPUser           string
	SMTPPassword       string
	EmailFrom          string
	HTTPPort           string
	LogLevel           string
	LogFormat          string
	SysadminEmail      string
	SysadminTenantName string

	// Time/Numeric
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	TokenTTL        time.Duration
	SMTPPort        int
}

// Log outputs the current configuration values at the info level.
func (c *Config) Log(ctx context.Context) {
	slog.InfoContext(ctx, "configuration details",
		"DatabaseURL", c.DatabaseURL,
		"RedisAddr", c.RedisAddr,
		"RedisPassword", c.RedisPassword,
		"PasetoSecretKey", c.PasetoSecretKey,
		"SMTPHost", c.SMTPHost,
		"SMTPUser", c.SMTPUser,
		"SMTPPassword", c.SMTPPassword,
		"EmailFrom", c.EmailFrom,
		"HTTPPort", c.HTTPPort,
		"LogLevel", c.LogLevel,
		"LogFormat", c.LogFormat,
		"SysadminEmail", c.SysadminEmail,
		"SysadminTenantName", c.SysadminTenantName,
		"ReadTimeout", c.ReadTimeout,
		"WriteTimeout", c.WriteTimeout,
		"ShutdownTimeout", c.ShutdownTimeout,
		"TokenTTL", c.TokenTTL,
		"SMTPPort", c.SMTPPort,
	)
}

// Load reads environment variables and returns a populated Config.
// It panics if any required variable is missing or malformed.
func Load() *Config {
	return &Config{
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		ReadTimeout:     getDurationEnv("READ_TIMEOUT", "10s"),
		WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", "30s"),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", "15s"),

		DatabaseURL: getRequiredEnv("DATABASE_URL"),

		RedisAddr:     getRequiredEnv("REDIS_ADDR"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		PasetoSecretKey: getRequiredEnv("PASETO_SECRET_KEY"),
		TokenTTL:        getDurationEnv("TOKEN_TTL", "24h"),

		SMTPHost:     getRequiredEnv("SMTP_HOST"),
		SMTPPort:     getIntEnv("SMTP_PORT", "587"),
		SMTPUser:     getRequiredEnv("SMTP_USER"),
		SMTPPassword: getRequiredEnv("SMTP_PASSWORD"),
		EmailFrom:    getRequiredEnv("EMAIL_FROM"),

		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		SysadminEmail:      getRequiredEnv("SYSADMIN_EMAIL"),
		SysadminTenantName: getEnv("SYSADMIN_TENANT_NAME", "System"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getRequiredEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("config: %s is required", key))
	}
	return value
}

func getDurationEnv(key, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	d, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Sprintf("config: %s is not a valid duration: %v", key, err))
	}
	return d
}

func getIntEnv(key, defaultValue string) int {
	value := getEnv(key, defaultValue)
	i, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("config: %s is not a valid integer: %v", key, err))
	}
	return i
}

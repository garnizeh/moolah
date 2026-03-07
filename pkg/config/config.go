package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Database
	DatabaseURL string

	// Redis
	RedisAddr     string
	RedisPassword string

	// PASETO
	PasetoSecretKey string // 32-byte hex-encoded symmetric key

	// SMTP / Mailer
	SMTPHost     string
	SMTPUser     string
	SMTPPassword string
	EmailFrom    string

	// Logging
	LogLevel  string // debug | info | warn | error
	LogFormat string // json | text

	// Server
	HTTPPort        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	// PASETO TTL
	TokenTTL time.Duration

	// Mailer Port
	SMTPPort int
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

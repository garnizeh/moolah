package log

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitWithWriter_Development(t *testing.T) {
	t.Setenv("APP_ENV", "development")

	// Save original logger and restore after test
	old := slog.Default()
	t.Cleanup(func() { slog.SetDefault(old) })

	var buf bytes.Buffer
	InitWithWriter(&buf)

	slog.Info("test message", "key", "value")

	output := buf.String()
	// Text handler format contains level=INFO and msg="test message"
	assert.Contains(t, output, "level=INFO")
	assert.Contains(t, output, "msg=\"test message\"")
	assert.Contains(t, output, "key=value")
}

func TestInitWithWriter_Production(t *testing.T) {
	t.Setenv("APP_ENV", "production")

	// Save original logger and restore after test
	old := slog.Default()
	t.Cleanup(func() { slog.SetDefault(old) })

	var buf bytes.Buffer
	InitWithWriter(&buf)

	slog.Info("prod message", "key", "value")

	output := buf.String()
	// JSON handler format
	assert.Contains(t, output, "\"level\":\"INFO\"")
	assert.Contains(t, output, "\"msg\":\"prod message\"")
	assert.Contains(t, output, "\"key\":\"value\"")
}

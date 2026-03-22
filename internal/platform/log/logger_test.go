package log

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitWithWriter_Development(t *testing.T) {
	os.Setenv("APP_ENV", "development")
	defer os.Unsetenv("APP_ENV")

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
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	var buf bytes.Buffer
	InitWithWriter(&buf)

	slog.Info("prod message", "key", "value")

	output := buf.String()
	// JSON handler format
	assert.Contains(t, output, "\"level\":\"INFO\"")
	assert.Contains(t, output, "\"msg\":\"prod message\"")
	assert.Contains(t, output, "\"key\":\"value\"")
}


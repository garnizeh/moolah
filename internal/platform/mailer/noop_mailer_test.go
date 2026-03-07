package mailer_test

import (
	"context"
	"testing"

	"github.com/garnizeh/moolah/internal/platform/mailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoopMailer(t *testing.T) {
	t.Parallel()

	got := mailer.NewNoopMailer()
	assert.NotNil(t, got)
}

func TestNoopMailer_SendOTP(t *testing.T) {
	t.Parallel()

	n := mailer.NewNoopMailer()
	err := n.SendOTP(context.Background(), "test@example.com", "123456")
	require.NoError(t, err)
}

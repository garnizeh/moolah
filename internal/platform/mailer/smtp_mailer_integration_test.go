//go:build integration

package mailer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/platform/mailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSMTPMailer_SendOTP_Integration(t *testing.T) {
	ctx := context.Background()

	// Define Mailhog container
	req := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog:v1.0.1",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForHTTP("/api/v2/messages").WithPort("8025/tcp"),
	}

	mailhog, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		_ = mailhog.Terminate(ctx)
	}()

	// Get mapped ports
	smtpHost, err := mailhog.Host(ctx)
	require.NoError(t, err)

	smtpPort, err := mailhog.MappedPort(ctx, "1025")
	require.NoError(t, err)

	httpPort, err := mailhog.MappedPort(ctx, "8025")
	require.NoError(t, err)

	// Create Mailer
	m, err := mailer.NewSMTPMailer(smtpHost, smtpPort.Int(), "", "", "noreply@moolah.test")
	require.NoError(t, err)

	// Send OTP
	to := "user@example.test"
	code := "123456"
	err = m.SendOTP(ctx, to, code)
	require.NoError(t, err)

	// Verify in Mailhog API
	// Give it a tiny bit of time to process
	time.Sleep(500 * time.Millisecond)

	apiURL := fmt.Sprintf("http://%s:%d/api/v2/messages", smtpHost, httpPort.Int())
	resp, err := http.Get(apiURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Items []struct {
			Content struct {
				Body string `json:"Body"`
			} `json:"Content"`
			Raw struct {
				To []string `json:"To"`
			} `json:"Raw"`
		} `json:"items"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	require.NotEmpty(t, result.Items, "No emails found in Mailhog")

	// Check the latest message
	found := false
	for _, item := range result.Items {
		if item.Raw.To[0] == to && assert.Contains(t, item.Content.Body, code) {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected email with code %s to %s not found", code, to)
}

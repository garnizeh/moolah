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
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMTPMailer_SendOTP_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Get Mailhog server using testutil helper
	mh := containers.NewMailhogServer(t)
	host, port := mh.SMTPHostAndPort(t)

	// Create Mailer
	m, err := mailer.NewSMTPMailer(host, port, "", "", "noreply@moolah.test")
	require.NoError(t, err)

	// Send OTP
	to := "user@example.test"
	code := "123456"
	err = m.SendOTP(ctx, to, code)
	require.NoError(t, err)

	// Verify in Mailhog API
	// Give it a tiny bit of time to process
	time.Sleep(500 * time.Millisecond)

	host, port = mh.APIHostAndPort(t)
	apiURL := fmt.Sprintf("http://%s:%d/api/v2/messages", host, port)
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

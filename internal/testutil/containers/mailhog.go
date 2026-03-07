//go:build integration

package containers

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestMailhog holds the exposed SMTP and HTTP API addresses of the Mailhog container.
type TestMailhog struct {
	SMTPAddr string // e.g. "localhost:1025"
	APIAddr  string // e.g. "http://localhost:8025"
}

// SMTPHostAndPort splits the SMTPAddr into host and port components. It fails the test if SMTPAddr is malformed.
func (m *TestMailhog) SMTPHostAndPort(t *testing.T) (string, int) {
	t.Helper()

	host, portStr, err := net.SplitHostPort(m.SMTPAddr)
	require.NoErrorf(t, err, "failed to split SMTP address: %s", m.SMTPAddr)

	port, err := strconv.Atoi(portStr)
	require.NoErrorf(t, err, "invalid port in SMTP address: %s", portStr)

	return host, port
}

// APIHostAndPort splits the APIAddr into host and port components. It fails the test if APIAddr is malformed.
func (m *TestMailhog) APIHostAndPort(t *testing.T) (string, int) {
	t.Helper()

	// APIAddr is expected to be in the format "http://host:port"
	trimmed := m.APIAddr
	if after, ok :=strings.CutPrefix(trimmed, "http://"); ok  {
		trimmed = after
	} else if after0, ok0 :=strings.CutPrefix(trimmed, "https://"); ok0  {
		trimmed = after0
	}

	host, portStr, err := net.SplitHostPort(trimmed)
	require.NoErrorf(t, err, "failed to split API address: %s", m.APIAddr)

	port, err := strconv.Atoi(portStr)
	require.NoErrorf(t, err, "invalid port in API address: %s", portStr)

	return host, port
}

// NewMailhogServer starts an ephemeral Mailhog container and returns the
// TestMailhog handle. Container is cleaned up when t completes.
func NewMailhogServer(t *testing.T) *TestMailhog {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog:v1.0.1",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForHTTP("/api/v2/messages").WithPort("8025/tcp"),
	}

	mailhogC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start mailhog container")

	smtpMappedPort, err := mailhogC.MappedPort(ctx, "1025")
	require.NoError(t, err, "failed to get mailhog smtp port")

	apiMappedPort, err := mailhogC.MappedPort(ctx, "8025")
	require.NoError(t, err, "failed to get mailhog api port")

	host, err := mailhogC.Host(ctx)
	require.NoError(t, err, "failed to get mailhog host")

	t.Cleanup(func() {
		err := mailhogC.Terminate(ctx)
		require.NoError(t, err, "failed to terminate mailhog container")
	})

	return &TestMailhog{
		SMTPAddr: fmt.Sprintf("%s:%s", host, smtpMappedPort.Port()),
		APIAddr:  fmt.Sprintf("http://%s:%s", host, apiMappedPort.Port()),
	}
}

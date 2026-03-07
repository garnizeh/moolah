package mailer

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMTPMailer_SendOTP_HappyPath(t *testing.T) {
	t.Parallel()

	// Start a local fake SMTP server
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	addr := l.Addr().String()
	_, portStr, _ := net.SplitHostPort(addr)
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)

	go func() {
		conn, acceptErr := l.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()

		writer := textproto.NewWriter(bufio.NewWriter(conn))
		reader := textproto.NewReader(bufio.NewReader(conn))

		_ = writer.PrintfLine("220 localhost ESMTP")
		for {
			line, readLineErr := reader.ReadLine()
			if readLineErr != nil {
				break
			}
			if strings.HasPrefix(line, "QUIT") {
				_ = writer.PrintfLine("221 Bye")
				break
			}
			if strings.HasPrefix(line, "DATA") {
				_ = writer.PrintfLine("354 Start mail input; end with <CRLF>.<CRLF>")
				_, _ = reader.ReadDotBytes()
				_ = writer.PrintfLine("250 OK")
				continue
			}
			_ = writer.PrintfLine("250 OK")
		}
	}()

	m, _ := NewSMTPMailer("127.0.0.1", port, "", "", "no-reply@moolah.io")
	err = m.SendOTP(context.Background(), "user@example.com", "123456")
	assert.NoError(t, err)
}

func TestNewSMTPMailer(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		m, err := NewSMTPMailer("localhost", 1025, "user", "pass", "no-reply@moolah.io")
		require.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("missing host", func(t *testing.T) {
		t.Parallel()
		_, err := NewSMTPMailer("", 1025, "user", "pass", "no-reply@moolah.io")
		assert.Error(t, err)
	})

	t.Run("missing from address", func(t *testing.T) {
		t.Parallel()
		_, err := NewSMTPMailer("localhost", 1025, "user", "pass", "")
		assert.Error(t, err)
	})
}

func TestNoopMailer(t *testing.T) {
	t.Parallel()
	m := NewNoopMailer()
	err := m.SendOTP(context.Background(), "test@user.com", "123456")
	assert.NoError(t, err)
}

func TestSMTPMailer_Validation(t *testing.T) {
	t.Parallel()
	m, _ := NewSMTPMailer("localhost", 1025, "", "", "no-reply@moolah.io")

	t.Run("missing recipient", func(t *testing.T) {
		t.Parallel()
		err := m.SendOTP(context.Background(), "", "123456")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "recipient address is required")
	})

	t.Run("missing code", func(t *testing.T) {
		t.Parallel()
		err := m.SendOTP(context.Background(), "user@example.com", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "OTP code is required")
	})
}

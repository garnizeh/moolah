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
	_, portStr, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	require.NoError(t, err)

	go func() {
		conn, acceptErr := l.Accept()
		if acceptErr != nil {
			return
		}
		defer func() {
			if closeErr := conn.Close(); closeErr != nil {
				t.Errorf("failed to close fake SMTP connection: %v", closeErr)
			}
		}()

		bufWriter := bufio.NewWriter(conn)
		writer := textproto.NewWriter(bufWriter)
		reader := textproto.NewReader(bufio.NewReader(conn))

		if lerr := writer.PrintfLine("220 localhost ESMTP"); lerr != nil {
			t.Errorf("failed to write greeting: %v", lerr)
			return
		}
		for {
			line, readLineErr := reader.ReadLine()
			if readLineErr != nil {
				break
			}
			if strings.HasPrefix(line, "QUIT") {
				if lerr := writer.PrintfLine("221 Bye"); lerr != nil {
					t.Errorf("failed to write bye: %v", lerr)
				}
				break
			}
			if strings.HasPrefix(line, "DATA") {
				if lerr := writer.PrintfLine("354 Start mail input; end with <CRLF>.<CRLF>"); lerr != nil {
					t.Errorf("failed to write data response: %v", lerr)
					return
				}
				if _, rerr := reader.ReadDotBytes(); rerr != nil {
					t.Errorf("failed to read data dot bytes: %v", rerr)
					return
				}
				if lerr := writer.PrintfLine("250 OK"); lerr != nil {
					t.Errorf("failed to write data 250 ok: %v", lerr)
					return
				}
				continue
			}
			if lerr := writer.PrintfLine("250 OK"); lerr != nil {
				t.Errorf("failed to write default 250 ok: %v", lerr)
				return
			}
		}
	}()

	m, err := NewSMTPMailer("127.0.0.1", port, "", "", "no-reply@moolah.io")
	require.NoError(t, err)
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

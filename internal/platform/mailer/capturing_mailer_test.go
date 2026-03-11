package mailer_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/garnizeh/moolah/internal/platform/mailer"
	"github.com/stretchr/testify/require"
)

func TestCapturingMailer_SendAndRetrieveOTP(t *testing.T) {
	t.Parallel()
	m := mailer.NewCapturingMailer()
	email := "test@example.com"
	err := m.SendOTP(context.Background(), email, "123456")
	require.NoError(t, err)
	got := m.OTPFor(email)
	require.Equal(t, "123456", got)
}

func TestCapturingMailer_ConcurrentSendOTP(t *testing.T) {
	t.Parallel()
	m := mailer.NewCapturingMailer()
	var wg sync.WaitGroup
	n := 100
	errs := make(chan error, n)

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			email := fmt.Sprintf("user%d@example.com", i)
			code := fmt.Sprintf("%06d", i)
			errs <- m.SendOTP(context.Background(), email, code)
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	for i := range n {
		email := fmt.Sprintf("user%d@example.com", i)
		expected := fmt.Sprintf("%06d", i)
		require.Equal(t, expected, m.OTPFor(email))
	}
}

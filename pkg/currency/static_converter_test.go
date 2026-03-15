package currency_test

import (
	"context"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/pkg/currency"
	"github.com/stretchr/testify/require"
)

func TestStaticConverter(t *testing.T) {
	t.Parallel()

	rates := map[string]map[string]int64{
		"USD": {
			"BRL": 50000, // 1 USD = 5.00 BRL
		},
		"EUR": {
			"USD": 11000, // 1 EUR = 1.10 USD
		},
	}

	converter := currency.NewStaticConverter(rates)
	ctx := context.Background()

	t.Run("same currency", func(t *testing.T) {
		t.Parallel()
		res, err := converter.Convert(ctx, 1000, "USD", "USD")
		require.NoError(t, err)
		require.Equal(t, int64(1000), res)
	})

	t.Run("defined rate: USD to BRL", func(t *testing.T) {
		t.Parallel()
		// (1000 * 50000) / 10000 = 5000
		res, err := converter.Convert(ctx, 1000, "USD", "BRL")
		require.NoError(t, err)
		require.Equal(t, int64(5000), res)
	})

	t.Run("defined rate: EUR to USD", func(t *testing.T) {
		t.Parallel()
		// (2000 * 11000) / 10000 = 2200
		res, err := converter.Convert(ctx, 2000, "EUR", "USD")
		require.NoError(t, err)
		require.Equal(t, int64(2200), res)
	})

	t.Run("missing rate: BRL to USD (fromCurrency exists but toCurrency missing)", func(t *testing.T) {
		t.Parallel()
		// USD exists in rates, but USD specifically to EUR is not defined.
		_, err := converter.Convert(ctx, 1000, "USD", "EUR")
		require.ErrorIs(t, err, domain.ErrRateNotFound)
	})

	t.Run("missing fromCurrency (entire map key missing)", func(t *testing.T) {
		t.Parallel()
		_, err := converter.Convert(ctx, 1000, "JPY", "USD")
		require.ErrorIs(t, err, domain.ErrRateNotFound)
	})

	t.Run("nil rates", func(t *testing.T) {
		t.Parallel()
		nilConverter := currency.NewStaticConverter(nil)
		_, err := nilConverter.Convert(ctx, 1000, "USD", "BRL")
		require.ErrorIs(t, err, domain.ErrRateNotFound)
	})
}

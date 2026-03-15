package currency_test

import (
	"context"
	"testing"

	"github.com/garnizeh/moolah/pkg/currency"
	"github.com/stretchr/testify/require"
)

func TestNoopConverter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	converter := currency.NewNoopConverter()

	t.Run("converts same currency", func(t *testing.T) {
		t.Parallel()
		res, err := converter.Convert(ctx, 1000, "USD", "USD")
		require.NoError(t, err)
		require.Equal(t, int64(1000), res)
	})

	t.Run("converts different currency", func(t *testing.T) {
		t.Parallel()
		res, err := converter.Convert(ctx, 1000, "USD", "BRL")
		require.NoError(t, err)
		require.Equal(t, int64(1000), res)
	})
}

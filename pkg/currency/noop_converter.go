package currency

import (
	"context"

	"github.com/garnizeh/moolah/internal/domain"
)

var _ domain.CurrencyConverter = (*NoopConverter)(nil)

// NoopConverter assumes all currencies are the same and returns the input unchanged.
type NoopConverter struct{}

// NewNoopConverter returns a new NoopConverter instance.
func NewNoopConverter() *NoopConverter {
	return &NoopConverter{}
}

// Convert returns amountCents unchanged.
func (c *NoopConverter) Convert(ctx context.Context, amountCents int64, fromCurrency, toCurrency string) (int64, error) {
	return amountCents, nil
}

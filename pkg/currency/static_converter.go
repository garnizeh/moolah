package currency

import (
	"context"
	"fmt"
	"sync"

	"github.com/garnizeh/moolah/internal/domain"
)

var _ domain.CurrencyConverter = (*StaticConverter)(nil)

// StaticConverter holds a static rate table for currency conversion.
// Rates are expressed in basis points: rate["USD"]["BRL"] = 50000 means 1 USD = 5.0000 BRL.
// The conversion formula is: (amountCents * rate) / 10000.
type StaticConverter struct {
	rates map[string]map[string]int64
	mu    sync.RWMutex
}

// NewStaticConverter returns a new StaticConverter with the given rates.
func NewStaticConverter(rates map[string]map[string]int64) *StaticConverter {
	if rates == nil {
		rates = make(map[string]map[string]int64)
	}
	return &StaticConverter{
		rates: rates,
	}
}

// Convert converts amountCents from fromCurrency to toCurrency using the static rate table.
// It uses integer arithmetic with a fixed precision of 4 decimal places (10000 divisor).
func (c *StaticConverter) Convert(ctx context.Context, amountCents int64, fromCurrency, toCurrency string) (int64, error) {
	if fromCurrency == toCurrency {
		return amountCents, nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	targetRates, ok := c.rates[fromCurrency]
	if !ok {
		return 0, fmt.Errorf("%w: %s to %s", domain.ErrRateNotFound, fromCurrency, toCurrency)
	}

	rate, ok := targetRates[toCurrency]
	if !ok {
		return 0, fmt.Errorf("%w: %s to %s", domain.ErrRateNotFound, fromCurrency, toCurrency)
	}

	// amountCents * rate / 10000 (basis points)
	// Example: 100 USD cents * 50000 (rate) / 10000 = 500 BRL cents.
	return (amountCents * rate) / 10000, nil
}

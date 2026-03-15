package domain

import (
	"context"
	"errors"
)

// ErrRateNotFound is returned when a currency conversion rate is not available.
var ErrRateNotFound = errors.New("currency conversion rate not found")

// CurrencyConverter converts a monetary amount from one currency to another.
// All amounts are expressed in cents (int64) to avoid float imprecision.
// Implementations must be safe for concurrent use.
type CurrencyConverter interface {
	// Convert returns amountCents in the target currency.
	// Returns ErrRateNotFound if the conversion rate is not available.
	Convert(ctx context.Context, amountCents int64, fromCurrency, toCurrency string) (int64, error)
}

package domain

import (
	"encoding/json"
	"time"
)

// Currency represents a monetary system (e.g., USD, BRL, BTC).
type Currency struct {
	ID               string          `json:"id"`
	Code             string          `json:"code"`
	Symbol           string          `json:"symbol"`
	FallbackDecimals int             `json:"fallback_decimals"`
	Config           json.RawMessage `json:"config"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

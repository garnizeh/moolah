package domain

import (
	"encoding/json"
	"time"
)

// Account represents a financial account belonging to an entity.
type Account struct {
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `json:"deleted_at,omitempty"`
	ID           string          `json:"id"`
	EntityID     string          `json:"entity_id"`
	CurrencyID   string          `json:"currency_id"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Metadata     json.RawMessage `json:"metadata"`
	BalanceCents int64           `json:"balance_cents"`
}

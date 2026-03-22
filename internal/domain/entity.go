package domain

import (
	"encoding/json"
	"time"
)

// Entity represents a family member or a cost center.
type Entity struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Role      string          `json:"role"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty"`
}

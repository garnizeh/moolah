package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategory_IsRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		parentID string
		want     bool
	}{
		{"root category (empty string)", "", true},
		{"child category (has ID)", "ulid_parent_123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Category{ParentID: tt.parentID}
			assert.Equal(t, tt.want, c.IsRoot())
		})
	}
}

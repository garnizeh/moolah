// Package ulid provides a thread-safe monotonic ULID generator.
package ulid

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	mu      sync.Mutex
	entropy = ulid.Monotonic(rand.Reader, 0)
)

// New returns a new ULID string. Safe for concurrent use.
// It will produce monotonically increasing identifiers when called within the same millisecond.
func New() string {
	mu.Lock()
	defer mu.Unlock()
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return id.String()
}

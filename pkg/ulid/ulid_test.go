package ulid

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("returns valid format", func(t *testing.T) {
		t.Parallel()
		id := New()
		assert.Len(t, id, 26, "ULID should be 26 characters")
	})

	t.Run("produces unique values", func(t *testing.T) {
		t.Parallel()
		const count = 1000
		ids := make(map[string]struct{}, count)
		for range count {
			id := New()
			if _, exists := ids[id]; exists {
				t.Errorf("Duplicate ULID generated: %s", id)
			}
			ids[id] = struct{}{}
		}
	})

	t.Run("is monotonic within same millisecond", func(t *testing.T) {
		t.Parallel()
		const count = 100
		ids := make([]string, count)
		for i := range count {
			ids[i] = New()
		}

		assert.True(t, sort.StringsAreSorted(ids), "ULIDs should be monotonically increasing")
	})

	t.Run("is thread-safe with race detector", func(t *testing.T) {
		t.Parallel()
		const count = 100
		const goroutines = 20
		var wg sync.WaitGroup
		wg.Add(goroutines)

		results := make(chan string, count*goroutines)

		for range goroutines {
			go func() {
				defer wg.Done()
				for range count {
					results <- New()
				}
			}()
		}

		wg.Wait()
		close(results)

		ids := make(map[string]struct{}, count*goroutines)
		for id := range results {
			if _, exists := ids[id]; exists {
				t.Errorf("Duplicate ULID generated in concurrent test: %s", id)
			}
			ids[id] = struct{}{}
		}
		assert.Len(t, ids, count*goroutines)
	})
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func BenchmarkNewParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			New()
		}
	})
}

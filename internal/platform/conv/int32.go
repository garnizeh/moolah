package conv

import (
	"fmt"
	"math"
)

// ToInt32 safely converts an int to an int32, returning an error if it overflows.
func ToInt32(n int) (int32, error) {
	if n < math.MinInt32 || n > math.MaxInt32 {
		return 0, fmt.Errorf("integer overflow: %d is out of range for int32", n)
	}
	return int32(n), nil
}

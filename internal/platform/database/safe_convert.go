package database

import (
	"fmt"
	"math"
)

// SafeInt64ToInt safely converts int64 to int, returning an error if the value overflows.
// On 64-bit platforms this is a no-op, but on 32-bit platforms int is 32 bits and
// large int64 values would silently overflow without this check.
func SafeInt64ToInt(v int64) (int, error) {
	if v > math.MaxInt || v < math.MinInt {
		return 0, fmt.Errorf("integer overflow: %d exceeds int range", v)
	}
	return int(v), nil
}

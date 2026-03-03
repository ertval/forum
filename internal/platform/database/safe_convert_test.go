package database

import (
	"math"
	"testing"
)

func TestSafeInt64ToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   int64
		want    int
		wantErr bool
	}{
		{"zero", 0, 0, false},
		{"positive", 42, 42, false},
		{"negative", -42, -42, false},
		{"max int", int64(math.MaxInt), math.MaxInt, false},
		{"min int", int64(math.MinInt), math.MinInt, false},
	}

	// On 32-bit platforms, these would overflow; on 64-bit they equal MaxInt/MinInt.
	if math.MaxInt < math.MaxInt64 {
		tests = append(tests,
			struct {
				name    string
				input   int64
				want    int
				wantErr bool
			}{"overflow positive", math.MaxInt64, 0, true},
			struct {
				name    string
				input   int64
				want    int
				wantErr bool
			}{"overflow negative", math.MinInt64, 0, true},
		)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SafeInt64ToInt(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("SafeInt64ToInt(%d) error = %v, wantErr %v", tc.input, err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("SafeInt64ToInt(%d) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

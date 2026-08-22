package numutil_test

import (
	"math"
	"testing"

	"github.com/vxcontrol/langchaingo/internal/numutil"
)

func TestSaturateInt32ClampsInsteadOfWrapping(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   int
		want int32
	}{
		{"above the ceiling", 3_000_000_000, math.MaxInt32},
		{"exactly the ceiling", math.MaxInt32, math.MaxInt32},
		{"one below the ceiling", math.MaxInt32 - 1, math.MaxInt32 - 1},
		{"below the floor", -3_000_000_000, math.MinInt32},
		{"ordinary", 16384, 16384},
		{"zero", 0, 0},
	} {
		if got := numutil.SaturateInt32(tc.in); got != tc.want {
			t.Errorf("%s: SaturateInt32(%d) = %d, want %d", tc.name, tc.in, got, tc.want)
		}
	}
}

package numutil

import "math"

func SaturateInt32(i int) int32 {
	switch {
	case i > math.MaxInt32:
		return math.MaxInt32
	case i < math.MinInt32:
		return math.MinInt32
	default:
		return int32(i)
	}
}

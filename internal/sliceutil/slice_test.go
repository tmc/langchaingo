package sliceutil

import (
	"testing"
)

func TestMinInt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		nums     []int
		expected int
	}{
		{
			name:     "ascending order",
			nums:     []int{1, 2, 3, 23, 34},
			expected: 1,
		},
		{
			name:     "mixed order",
			nums:     []int{3, 2, 1, 34, 2213},
			expected: 1,
		},
		{
			name:     "nil slice",
			nums:     nil,
			expected: 0,
		},
		{
			name:     "empty slice",
			nums:     []int{},
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MinInt(tc.nums)
			if got != tc.expected {
				t.Errorf("MinInt(%v) = %v, want %v", tc.nums, got, tc.expected)
			}
		})
	}
}

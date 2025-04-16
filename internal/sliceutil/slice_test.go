package sliceutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinInt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		nums     []int
		expected int
	}{
		{
			nums:     []int{1, 2, 3, 23, 34},
			expected: 1,
		},
		{
			nums:     []int{3, 2, 1, 34, 2213},
			expected: 1,
		},
		{
			nums:     nil,
			expected: 0,
		},
		{
			nums:     []int{},
			expected: 0,
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expected, MinInt(tc.nums))
	}
}

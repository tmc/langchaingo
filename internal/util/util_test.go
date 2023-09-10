package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	t.Parallel()
	type testCaseDef struct {
		name           string
		sourceMap      map[string]any
		expectedNewMap map[string]any
	}

	testCases := []testCaseDef{
		{
			name:           "nil map becomes empty map",
			sourceMap:      nil,
			expectedNewMap: map[string]any{},
		},
		{
			name:           "empty map is preserved",
			sourceMap:      map[string]any{},
			expectedNewMap: map[string]any{},
		},
		{
			name:           "simple map with nil value is preserved",
			sourceMap:      map[string]any{"a": 1, "b": 3.1415926, "c": "hi", "d": nil},
			expectedNewMap: map[string]any{"a": 1, "b": 3.1415926, "c": "hi", "d": nil},
		},
	}

	for _, tc := range testCases {
		tc := tc // see paralleltest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			newMap := NewMap(tc.sourceMap)
			assert.Equal(t, tc.expectedNewMap, newMap)
		})
	}
}

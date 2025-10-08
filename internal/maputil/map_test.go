package maputil

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: []string{},
		},
		{
			name: "single key",
			input: map[string]any{
				"key1": "value1",
			},
			expected: []string{"key1"},
		},
		{
			name: "multiple keys",
			input: map[string]any{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			expected: []string{"key1", "key2", "key3"},
		},
		{
			name: "nil values",
			input: map[string]any{
				"key1": nil,
				"key2": nil,
			},
			expected: []string{"key1", "key2"},
		},
		{
			name: "various types",
			input: map[string]any{
				"string": "hello",
				"int":    123,
				"float":  3.14,
				"bool":   false,
				"slice":  []int{1, 2, 3},
				"map":    map[string]string{"nested": "value"},
			},
			expected: []string{"string", "int", "float", "bool", "slice", "map"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListKeys(tt.input)

			// Sort both slices for comparison since map iteration order is not guaranteed
			sort.Strings(result)
			sort.Strings(tt.expected)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListKeys_TypedMaps(t *testing.T) {
	t.Run("string map", func(t *testing.T) {
		m := map[string]string{
			"a": "alpha",
			"b": "beta",
			"c": "gamma",
		}
		keys := ListKeys(m)
		assert.Len(t, keys, 3)

		sort.Strings(keys)
		assert.Equal(t, []string{"a", "b", "c"}, keys)
	})

	t.Run("int map", func(t *testing.T) {
		m := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		keys := ListKeys(m)
		assert.Len(t, keys, 3)
		assert.Contains(t, keys, "one")
		assert.Contains(t, keys, "two")
		assert.Contains(t, keys, "three")
	})

	t.Run("struct map", func(t *testing.T) {
		type testStruct struct {
			Name  string
			Value int
		}

		m := map[string]testStruct{
			"first":  {Name: "First", Value: 1},
			"second": {Name: "Second", Value: 2},
		}
		keys := ListKeys(m)
		assert.Len(t, keys, 2)
		assert.Contains(t, keys, "first")
		assert.Contains(t, keys, "second")
	})
}

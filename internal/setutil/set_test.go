package setutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSet(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]struct{}
	}{
		{
			name:     "empty list",
			input:    []string{},
			expected: map[string]struct{}{},
		},
		{
			name:  "single element",
			input: []string{"a"},
			expected: map[string]struct{}{
				"a": {},
			},
		},
		{
			name:  "multiple unique elements",
			input: []string{"a", "b", "c"},
			expected: map[string]struct{}{
				"a": {},
				"b": {},
				"c": {},
			},
		},
		{
			name:  "duplicate elements",
			input: []string{"a", "b", "a", "c", "b"},
			expected: map[string]struct{}{
				"a": {},
				"b": {},
				"c": {},
			},
		},
		{
			name:  "empty strings",
			input: []string{"", "a", ""},
			expected: map[string]struct{}{
				"":  {},
				"a": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSet(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDifference(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		set      map[string]struct{}
		expected []string
	}{
		{
			name:     "empty list and set",
			list:     []string{},
			set:      map[string]struct{}{},
			expected: []string{},
		},
		{
			name:     "empty list",
			list:     []string{},
			set:      map[string]struct{}{"a": {}},
			expected: []string{},
		},
		{
			name:     "empty set",
			list:     []string{"a", "b"},
			set:      map[string]struct{}{},
			expected: []string{"a", "b"},
		},
		{
			name: "no difference",
			list: []string{"a", "b"},
			set: map[string]struct{}{
				"a": {},
				"b": {},
			},
			expected: []string{},
		},
		{
			name: "partial difference",
			list: []string{"a", "b", "c"},
			set: map[string]struct{}{
				"b": {},
			},
			expected: []string{"a", "c"},
		},
		{
			name: "complete difference",
			list: []string{"a", "b", "c"},
			set: map[string]struct{}{
				"x": {},
				"y": {},
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "duplicates in list",
			list: []string{"a", "b", "a", "c", "b"},
			set: map[string]struct{}{
				"b": {},
			},
			expected: []string{"a", "a", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Difference(tt.list, tt.set)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		set      map[string]struct{}
		expected []string
	}{
		{
			name:     "empty list and set",
			list:     []string{},
			set:      map[string]struct{}{},
			expected: []string{},
		},
		{
			name:     "empty list",
			list:     []string{},
			set:      map[string]struct{}{"a": {}},
			expected: []string{},
		},
		{
			name:     "empty set",
			list:     []string{"a", "b"},
			set:      map[string]struct{}{},
			expected: []string{},
		},
		{
			name: "complete intersection",
			list: []string{"a", "b"},
			set: map[string]struct{}{
				"a": {},
				"b": {},
				"c": {},
			},
			expected: []string{"a", "b"},
		},
		{
			name: "partial intersection",
			list: []string{"a", "b", "c"},
			set: map[string]struct{}{
				"b": {},
				"d": {},
			},
			expected: []string{"b"},
		},
		{
			name: "no intersection",
			list: []string{"a", "b", "c"},
			set: map[string]struct{}{
				"x": {},
				"y": {},
			},
			expected: []string{},
		},
		{
			name: "duplicates in list",
			list: []string{"a", "b", "a", "c", "b"},
			set: map[string]struct{}{
				"a": {},
				"b": {},
			},
			expected: []string{"a", "b", "a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersection(tt.list, tt.set)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetOperationsCombined(t *testing.T) {
	// Test that ToSet and set operations work together correctly
	list1 := []string{"a", "b", "c", "d"}
	list2 := []string{"c", "d", "e", "f"}

	set2 := ToSet(list2)

	// Elements in list1 but not in list2
	diff := Difference(list1, set2)
	assert.Equal(t, []string{"a", "b"}, diff)

	// Elements in both lists
	inter := Intersection(list1, set2)
	assert.Equal(t, []string{"c", "d"}, inter)
}

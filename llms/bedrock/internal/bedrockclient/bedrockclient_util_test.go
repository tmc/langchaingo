package bedrockclient

import (
	"testing"
)

type ValidStruct struct {
	Field1 string `document:"field1"`
	Field2 int    `document:"field2"`
}

type InvalidStruct struct {
	Field1 string `document:"field1"`
	Field2 int    // missing document tag
}

type NestedValidStruct struct {
	Nested ValidStruct `document:"nested"`
	Value  string      `document:"value"`
}

type NestedInvalidStruct struct {
	Nested InvalidStruct `document:"nested"`
	Value  string        `document:"value"`
}

type StructWithSlice struct {
	Items []string `document:"items"`
	Name  string   `document:"name"`
}

type StructWithMap struct {
	Data map[string]int `document:"data"`
	Name string         `document:"name"`
}

func TestIsSmithyValidObject(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		// Simple types
		{"nil", nil, true},
		{"string", "test", true},
		{"int", 42, true},
		{"bool", true, true},
		{"float64", 3.14, true},

		// Valid structs
		{"valid struct", ValidStruct{Field1: "test", Field2: 42}, true},
		{"valid struct pointer", &ValidStruct{Field1: "test", Field2: 42}, true},
		{"nil pointer", (*ValidStruct)(nil), true},
		{"nested valid struct", NestedValidStruct{Nested: ValidStruct{}, Value: "test"}, true},

		// Invalid structs
		{"invalid struct", InvalidStruct{Field1: "test", Field2: 42}, false},
		{"invalid struct pointer", &InvalidStruct{Field1: "test", Field2: 42}, false},
		{"nested invalid struct", NestedInvalidStruct{Nested: InvalidStruct{}, Value: "test"}, false},

		// Valid slices
		{"empty slice", []string{}, true},
		{"slice of simple types", []string{"test1", "test2", "test3"}, true},
		{"slice of integers", []int{1, 2, 3}, true},
		{"slice of valid structs", []ValidStruct{{Field1: "a", Field2: 1}, {Field1: "b", Field2: 2}}, true},
		{"slice of valid struct pointers", []*ValidStruct{{Field1: "a", Field2: 1}, {Field1: "b", Field2: 2}}, true},
		{"slice with nil pointers", []*ValidStruct{nil, {Field1: "test", Field2: 42}}, true},

		// Invalid slices
		{"slice of invalid structs", []InvalidStruct{{Field1: "test", Field2: 42}}, false},
		{"slice of invalid struct pointers", []*InvalidStruct{{Field1: "test", Field2: 42}}, false},
		{"mixed slice with invalid struct", []any{ValidStruct{Field1: "test", Field2: 42}, InvalidStruct{Field1: "test", Field2: 42}}, false},

		// Valid maps
		{"empty map", map[string]int{}, true},
		{"map with simple types", map[string]int{"key1": 1, "key2": 2}, true},
		{"map with int keys", map[int]string{1: "one", 2: "two"}, true},
		{"map with valid struct values", map[string]ValidStruct{"a": {Field1: "test1", Field2: 1}, "b": {Field1: "test2", Field2: 2}}, true},
		{"map with valid struct pointer values", map[string]*ValidStruct{"a": {Field1: "test1", Field2: 1}, "b": nil}, true},

		// Invalid maps
		{"map with invalid struct values", map[string]InvalidStruct{"a": {Field1: "test", Field2: 42}}, false},
		{"map with complex values", map[string][]InvalidStruct{"key": {{Field1: "test", Field2: 42}}}, false},

		// Nested collections
		{"slice of slices", [][]string{{"a", "b"}, {"c", "d"}}, true},
		{"slice of maps", []map[string]int{{"key1": 1}, {"key2": 2}}, true},
		{"map of slices", map[string][]int{"key1": {1, 2}, "key2": {3, 4}}, true},
		{"map of maps", map[string]map[string]int{"outer": {"inner": 42}}, true},

		// Invalid nested collections
		{"map of invalid slices", map[string][]InvalidStruct{"key": {{Field1: "test", Field2: 42}}}, false},

		// Structs with collections
		{"struct with valid slice", StructWithSlice{Items: []string{"a", "b"}, Name: "test"}, true},
		{"struct with valid map", StructWithMap{Data: map[string]int{"key": 42}, Name: "test"}, true},

		// Non-simple, non-struct, non-collection types
		{"channel", make(chan int), false},
		{"function", func() {}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSmithyValidObject(tt.input)
			if result != tt.expected {
				t.Errorf("isSmithyValidObject(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

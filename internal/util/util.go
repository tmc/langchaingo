// Package util contins general helper functions.
package util

// ToSet converts a list to a set.
func ToSet(list []string) map[string]struct{} {
	set := make(map[string]struct{}, 0)
	for _, v := range list {
		set[v] = struct{}{}
	}
	return set
}

// Difference returns the elements in list that are not in set.
func Difference(list []string, set map[string]struct{}) []string {
	diff := make([]string, 0)
	for _, v := range list {
		if _, ok := set[v]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}

// Intersection returns the elements in list that are in set.
func Intersection(list []string, set map[string]struct{}) []string {
	intersection := make([]string, 0)
	for _, v := range list {
		if _, ok := set[v]; ok {
			intersection = append(intersection, v)
		}
	}
	return intersection
}

func ListKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// NewMap will create a new map based upon the provided map.
// If the provided map is `nil`, a new, empty map will be returned.
func NewMap[K, V comparable](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

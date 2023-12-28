// Package util contains general helper functions.
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

// MinInt returns the minimum value in nums.
// If nums is empty, it returns 0.
func MinInt(nums []int) int {
	var min int
	for idx := 0; idx < len(nums); idx++ {
		item := nums[idx]
		if idx == 0 {
			min = item
			continue
		}
		if item < min {
			min = item
		}
	}
	return min
}

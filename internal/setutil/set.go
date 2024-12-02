package setutil

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

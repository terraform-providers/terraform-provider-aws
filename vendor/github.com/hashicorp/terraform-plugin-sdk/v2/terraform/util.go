package terraform

import (
	"sort"
)

// deduplicate a slice of strings
func uniqueStrings(s []string) []string {
	if len(s) < 2 {
		return s
	}

	sort.Strings(s)
	result := make([]string, 1, len(s))
	result[0] = s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != result[len(result)-1] {
			result = append(result, s[i])
		}
	}
	return result
}

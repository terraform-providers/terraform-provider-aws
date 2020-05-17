// +build !generate

package namevaluesfilters

import (
	"fmt"
)

// Custom EC2 filter functions.

// Ec2Tags creates NameValuesFilters from a map of keyvalue tags.
func Ec2Tags(tags map[string]string) NameValuesFilters {
	m := make(map[string]string, len(tags))

	for k, v := range tags {
		m[fmt.Sprintf("tag:%s", k)] = v
	}

	return New(m)
}

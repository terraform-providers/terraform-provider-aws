//go:generate go run -tags generate generators/servicefilters/main.go

package namevaluesfilters

// NameValuesFilters is a standard implementation for AWS resource filters.
// The AWS Go SDK is split into multiple service packages, each service with
// its own Go struct type representing a resource filter. To standardize logic
// across all these Go types, we convert them into this Go type.
type NameValuesFilters map[string][]string

// Map returns filter names mapped to their values.
func (filters NameValuesFilters) Map() map[string][]string {
	result := make(map[string][]string, len(filters))

	for k, v := range filters {
		result[k] = make([]string, len(v))
		copy(result[k], v)
	}

	return result
}

// Merge adds missing and updates existing filters.
func (filters NameValuesFilters) Merge(mergeFilters NameValuesFilters) NameValuesFilters {
	result := make(NameValuesFilters)

	for k, v := range filters {
		result[k] = v
	}

	for k, v := range mergeFilters {
		if values, ok := result[k]; ok {
			result[k] = append(values, v...)
		} else {
			result[k] = v
		}
	}

	return result
}

// New creates NameValuesFilters from common Terraform Provider SDK types.
// Supports map[string]string, map[string][]string, TODO.
func New(i interface{}) NameValuesFilters {
	switch value := i.(type) {
	case map[string]string:
		nvfm := make(NameValuesFilters, len(value))

		for k, v := range value {
			nvfm[k] = []string{v}
		}

		return nvfm
	case map[string][]string:
		return NameValuesFilters(value)
	default:
		return make(NameValuesFilters)
	}
}

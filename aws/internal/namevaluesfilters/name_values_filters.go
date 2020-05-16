//go:generate go run -tags generate generators/servicefilters/main.go

package namevaluesfilters

// NameValuesFilters is a standard implementation for AWS resource filters.
// The AWS Go SDK is split into multiple service packages, each service with
// its own Go struct type representing a resource filter. To standardize logic
// across all these Go types, we convert them into this Go type.
type NameValuesFilters map[string][]string

// Map returns filter names mapped to their values.
// Duplicate values are eliminated and empty values removed.
func (filters NameValuesFilters) Map() map[string][]string {
	result := make(map[string][]string)

	for k, v := range filters {
		targetValues := make([]string, 0)

	SOURCE_VALUES:
		for _, sourceValue := range v {
			if sourceValue == "" {
				continue
			}

			for _, targetValue := range targetValues {
				if sourceValue == targetValue {
					continue SOURCE_VALUES
				}
			}

			targetValues = append(targetValues, sourceValue)
		}

		if len(targetValues) == 0 {
			continue
		}

		result[k] = targetValues
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

//go:generate go run -tags generate generators/servicefilters/main.go

package namevaluesfilters

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

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
// Supports map[string]string, map[string][]string, *schema.Set.
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
	case *schema.Set:
		// The set of filters described by Schema().
		filters := value.List()
		nvfm := make(NameValuesFilters, len(filters))

		for _, filter := range filters {
			filterMap := filter.(map[string]interface{})
			name := filterMap["name"].(string)
			values := filterMap["values"].(*schema.Set)
			nvfm[name] = make([]string, values.Len())
			for _, value := range values.List() {
				nvfm[name] = append(nvfm[name], value.(string))
			}
		}

		return nvfm
	default:
		return make(NameValuesFilters)
	}
}

// Schema returns a *schema.Schema that represents a set of custom filtering criteria
// that a user can specify as input to a data source.
// It is conventional for an attribute of this type to be included as a top-level attribute called "filter".
func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},

				"values": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

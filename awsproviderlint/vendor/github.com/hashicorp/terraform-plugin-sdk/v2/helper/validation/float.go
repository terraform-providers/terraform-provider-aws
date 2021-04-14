package validation

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FloatBetween returns a SchemaValidateFunc which tests if the provided value
// is of type float64 and is between min and max (inclusive).
func FloatBetween(min, max float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(float64)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be float64", k))
			return
		}

		if v < min || v > max {
			es = append(es, fmt.Errorf("expected %s to be in the range (%f - %f), got %f", k, min, max, v))
			return
		}

		return
	}
}

// FloatAtLeast returns a SchemaValidateFunc which tests if the provided value
// is of type float and is at least min (inclusive)
func FloatAtLeast(min float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(float64)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be float", k))
			return
		}

		if v < min {
			es = append(es, fmt.Errorf("expected %s to be at least (%f), got %f", k, min, v))
			return
		}

		return
	}
}

// FloatAtMost returns a SchemaValidateFunc which tests if the provided value
// is of type float and is at most max (inclusive)
func FloatAtMost(max float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(float64)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be float", k))
			return
		}

		if v > max {
			es = append(es, fmt.Errorf("expected %s to be at most (%f), got %f", k, max, v))
			return
		}

		return
	}
}

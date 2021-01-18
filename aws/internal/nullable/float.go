package nullable

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	TypeNullableFloat = schema.TypeString
)

// Float represents a nullable floating-point value
type Float string

// IsNull returns true if the value represents a null value
func (f Float) IsNull() bool {
	return f == ""
}

// Value returns the following information about the variable:
// * value - Returns the value of the variable, or 0 if it is null or an error
// * null - Returns true if the variable is null
// * err - Returns an error from parsing the variable, if there is one
func (f Float) Value() (float64, bool, error) {
	if f.IsNull() {
		return 0, true, nil
	}

	value, err := strconv.ParseFloat(string(f), 64)
	if err != nil {
		return 0, false, err
	}
	return value, false, nil
}

// ValidateTypeStringNullableFloat provides custom error messaging for TypeString floats
// Some arguments require a floating-point value or unspecified, empty field.
func ValidateTypeStringNullableFloat(v interface{}, k string) (ws []string, es []error) {
	value, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := strconv.ParseFloat(value, 64); err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse '%s' as float: %w", k, value, err))
	}

	return
}

// ValidateFloatBetween returns a SchemaValidateFunc which tests if the provided value
// is parseable as a float and is between min and max (inclusive)
func ValidateFloatBetween(min, max float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (ws []string, es []error) {
		value, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if value == "" {
			return
		}

		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			es = append(es, fmt.Errorf("%s: cannot parse '%s' as float: %w", k, value, err))
			return
		}

		if v < min || v > max {
			es = append(es, fmt.Errorf("expected %s to be in the range (%f - %f), got %f", k, min, max, v))
		}

		return
	}
}

// ValidateFloatAtLeast returns a SchemaValidateFunc which tests if the provided value
// is parseable as a float and is at least min (inclusive)
func ValidateFloatAtLeast(min float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (ws []string, es []error) {
		value, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if value == "" {
			return
		}

		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			es = append(es, fmt.Errorf("%s: cannot parse '%s' as float: %w", k, value, err))
			return
		}

		if v < min {
			es = append(es, fmt.Errorf("expected %s to be at least (%f), got %f", k, min, v))
		}

		return
	}
}

// ValidateFloatAtMost returns a SchemaValidateFunc which tests if the provided value
// is parseable as a float and is at most max (inclusive)
func ValidateFloatAtMost(max float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (ws []string, es []error) {
		value, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if value == "" {
			return
		}

		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			es = append(es, fmt.Errorf("%s: cannot parse '%s' as float: %w", k, value, err))
			return
		}

		if v > max {
			es = append(es, fmt.Errorf("expected %s to be at most (%f), got %f", k, max, v))
		}

		return
	}
}

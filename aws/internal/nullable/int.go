package nullable

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	TypeNullableInt = schema.TypeString
)

// Int represents a nullable integer value
type Int string

// IsNull returns true if the value represents a null value
func (i Int) IsNull() bool {
	return i == ""
}

// Value returns the following information about the variable:
// * value - Returns the value of the variable, or 0 if it is null or an error
// * null - Returns true if the variable is null
// * err - Returns an error from parsing the variable, if there is one
func (i Int) Value() (int64, bool, error) {
	if i.IsNull() {
		return 0, true, nil
	}

	value, err := strconv.ParseInt(string(i), 10, 64)
	if err != nil {
		return 0, false, err
	}
	return value, false, nil
}

// ValidateTypeStringNullableInt provides custom error messaging for TypeString ints
// Some arguments require an int value or unspecified, empty field.
func ValidateTypeStringNullableInt(v interface{}, k string) (ws []string, es []error) {
	value, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %w", k, value, err))
	}

	return
}

// ValidateIntBetween returns a SchemaValidateFunc which tests if the provided value
// is parseable as an int and is between min and max (inclusive)
func ValidateIntBetween(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (ws []string, es []error) {
		value, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if value == "" {
			return
		}

		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %w", k, value, err))
			return
		}

		if v < int64(min) || v > int64(max) {
			es = append(es, fmt.Errorf("expected %s to be in the range (%d - %d), got %d", k, min, max, v))
		}

		return
	}
}

// ValidateIntAtLeast returns a SchemaValidateFunc which tests if the provided value
// is parseable as an int and is at least min (inclusive)
func ValidateIntAtLeast(min int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (ws []string, es []error) {
		value, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if value == "" {
			return
		}

		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %w", k, value, err))
			return
		}

		if v < int64(min) {
			es = append(es, fmt.Errorf("expected %s to be at least (%d), got %d", k, min, v))
		}

		return
	}
}

// ValidateIntAtMost returns a SchemaValidateFunc which tests if the provided value
// is parseable as an int and is at most max (inclusive)
func ValidateIntAtMost(max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (ws []string, es []error) {
		value, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if value == "" {
			return
		}

		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %w", k, value, err))
			return
		}

		if v > int64(max) {
			es = append(es, fmt.Errorf("expected %s to be at most (%d), got %d", k, max, v))
		}

		return
	}
}

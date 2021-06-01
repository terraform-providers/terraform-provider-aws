package nullable

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	TypeNullableInt = schema.TypeString
)

type Int string

func (i Int) IsNull() bool {
	return i == ""
}

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

// ValidateTypeStringNullableIntAtLeast provides custom error messaging for TypeString ints
// Some arguments require an int value or unspecified, empty field.
func ValidateTypeStringNullableIntAtLeast(min int) schema.SchemaValidateFunc {
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

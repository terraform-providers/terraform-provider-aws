package nullable

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	TypeNullableBool = schema.TypeString
)

type Bool string

func (b Bool) IsNull() bool {
	return b == ""
}

func (b Bool) Value() (bool, bool, error) {
	if b.IsNull() {
		return false, true, nil
	}

	value, err := strconv.ParseBool(string(b))
	if err != nil {
		return false, false, err
	}
	return value, false, nil
}

func NewBool(v bool) Bool {
	return Bool(strconv.FormatBool(v))
}

// ValidateTypeStringNullableBool provides custom error messaging for TypeString booleans
// Some arguments require a boolean value or unspecified, empty field.
func ValidateTypeStringNullableBool(v interface{}, k string) (ws []string, es []error) {
	value, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := strconv.ParseBool(value); err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse '%s' as boolean: %w", k, value, err))
	}

	return
}

// DiffSuppressNullableBoolFalseAsNull allows false to be treated equivalently to null.
// This can be used to allow a practitioner to set false when the API requires a null value,
// as a convenience.
func DiffSuppressNullableBoolFalseAsNull(k, o, n string, d *schema.ResourceData) bool {
	ov, onull, _ := Bool(o).Value()
	nv, nnull, _ := Bool(n).Value()
	if !ov && nnull || onull && !nv {
		return true
	}
	return false
}

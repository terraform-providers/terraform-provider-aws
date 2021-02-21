package tftypes

import "fmt"

// Map is a Terraform type representing an unordered collection of elements,
// all of the same type, each identifiable with a unique string key.
type Map struct {
	AttributeType Type
}

// Is returns whether `t` is a Map type or not. If `t` is an instance of the
// Map type and its AttributeType property is not nil, it will only return true
// if its AttributeType is considered the same type as `m`'s AttributeType.
func (m Map) Is(t Type) bool {
	v, ok := t.(Map)
	if !ok {
		return false
	}
	if v.AttributeType != nil {
		return m.AttributeType.Is(v.AttributeType)
	}
	return ok
}

func (m Map) String() string {
	return "tftypes.Map"
}

func (m Map) private() {}

// MarshalJSON returns a JSON representation of the full type signature of `m`,
// including its AttributeType.
//
// Deprecated: this is not meant to be called by third-party code.
func (m Map) MarshalJSON() ([]byte, error) {
	attributeType, err := m.AttributeType.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling tftypes.Map's attribute type %T to JSON: %w", m.AttributeType, err)
	}
	return []byte(`["map",` + string(attributeType) + `]`), nil
}

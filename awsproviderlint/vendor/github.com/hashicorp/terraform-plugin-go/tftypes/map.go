package tftypes

import (
	"fmt"
	"sort"
)

// Map is a Terraform type representing an unordered collection of elements,
// all of the same type, each identifiable with a unique string key.
type Map struct {
	AttributeType Type

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

// Equal returns true if the two Maps are exactly equal. Unlike Is, passing in
// a Map with no AttributeType will always return false.
func (m Map) Equal(o Map) bool {
	return m.equals(o, true)
}

// Is returns whether `t` is a Map type or not. If `t` is an instance of the
// Map type and its AttributeType property is not nil, it will only return true
// if its AttributeType is considered the same type as `m`'s AttributeType.
func (m Map) Is(t Type) bool {
	return m.equals(t, false)
}

func (m Map) equals(t Type, exact bool) bool {
	v, ok := t.(Map)
	if !ok {
		return false
	}
	if v.AttributeType == nil || m.AttributeType == nil {
		// when doing exact comparisons, we can't compare types that
		// don't have element types set, so we just consider them not
		// equal
		//
		// when doing inexact comparisons, the absence of an element
		// type just means "is this a Map?" We know it is, so return
		// true if and only if m has an ElementType and t doesn't. This
		// behavior only makes sense if the user is trying to see if a
		// proper type is a map, so we want to ensure that the method
		// receiver always has an element type.
		if exact {
			return false
		}
		return m.AttributeType != nil
	}
	return m.AttributeType.equals(v.AttributeType, exact)
}

func (m Map) String() string {
	return "tftypes.Map[" + m.AttributeType.String() + "]"
}

func (m Map) private() {}

func (m Map) supportedGoTypes() []string {
	return []string{"map[string]tftypes.Value"}
}

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

func valueFromMap(typ Type, in interface{}) (Value, error) {
	switch value := in.(type) {
	case map[string]Value:
		keys := make([]string, 0, len(value))
		for k := range value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var elType Type
		for _, k := range keys {
			v := value[k]
			if err := useTypeAs(v.Type(), typ, NewAttributePath().WithElementKeyString(k)); err != nil {
				return Value{}, err
			}
			if elType == nil {
				elType = v.Type()
			}
			if !elType.equals(v.Type(), true) {
				return Value{}, fmt.Errorf("maps must only contain one type of element, saw %s and %s", elType, v.Type())
			}
		}
		return Value{
			typ:   Map{AttributeType: typ},
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.Map; expected types are: %s", in, formattedSupportedGoTypes(Map{}))
	}
}

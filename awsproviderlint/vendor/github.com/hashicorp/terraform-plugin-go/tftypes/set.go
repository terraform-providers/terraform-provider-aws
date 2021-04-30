package tftypes

import (
	"fmt"
)

// Set is a Terraform type representing an unordered collection of unique
// elements, all of the same type.
type Set struct {
	ElementType Type

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

// Equal returns true if the two Sets are exactly equal. Unlike Is, passing in
// a Set with no ElementType will always return false.
func (s Set) Equal(o Set) bool {
	return s.equals(o, true)
}

// Is returns whether `t` is a Set type or not. If `t` is an instance of the
// Set type and its ElementType property is nil, it will return true. If `t`'s
// ElementType property is not nil, it will only return true if its ElementType
// is considered the same type as `s`'s ElementType.
func (s Set) Is(t Type) bool {
	return s.equals(t, false)
}

func (s Set) equals(t Type, exact bool) bool {
	v, ok := t.(Set)
	if !ok {
		return false
	}
	if v.ElementType == nil || s.ElementType == nil {
		// when doing exact comparisons, we can't compare types that
		// don't have element types set, so we just consider them not
		// equal
		//
		// when doing inexact comparisons, the absence of an element
		// type just means "is this a Set?" We know it is, so return
		// true if and only if s has an ElementType and t doesn't. This
		// behavior only makes sense if the user is trying to see if a
		// proper type is a set, so we want to ensure that the method
		// receiver always has an element type.
		if exact {
			return false
		}
		return s.ElementType != nil
	}
	return s.ElementType.equals(v.ElementType, exact)
}

func (s Set) String() string {
	return "tftypes.Set[" + s.ElementType.String() + "]"
}

func (s Set) private() {}

func (s Set) supportedGoTypes() []string {
	return []string{"[]tftypes.Value"}
}

func valueFromSet(typ Type, in interface{}) (Value, error) {
	switch value := in.(type) {
	case []Value:
		var elType Type
		for _, v := range value {
			if err := useTypeAs(v.Type(), typ, NewAttributePath().WithElementKeyValue(v)); err != nil {
				return Value{}, err
			}
			if elType == nil {
				elType = v.Type()
			}
			if !elType.equals(v.Type(), true) {
				return Value{}, fmt.Errorf("sets must only contain one type of element, saw %s and %s", elType, v.Type())
			}
		}
		return Value{
			typ:   Set{ElementType: typ},
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.Set; expected types are: %s", in, formattedSupportedGoTypes(Set{}))
	}
}

// MarshalJSON returns a JSON representation of the full type signature of `s`,
// including its ElementType.
//
// Deprecated: this is not meant to be called by third-party code.
func (s Set) MarshalJSON() ([]byte, error) {
	elementType, err := s.ElementType.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling tftypes.Set's element type %T to JSON: %w", s.ElementType, err)
	}
	return []byte(`["set",` + string(elementType) + `]`), nil
}

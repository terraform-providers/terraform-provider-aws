package tftypes

import (
	"fmt"
)

// List is a Terraform type representing an ordered collection of elements, all
// of the same type.
type List struct {
	ElementType Type

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

// Equal returns true if the two Lists are exactly equal. Unlike Is, passing in
// a List with no ElementType will always return false.
func (l List) Equal(o List) bool {
	return l.equals(o, true)
}

// Is returns whether `t` is a List type or not. If `t` is an instance of the
// List type and its ElementType property is nil, it will return true. If `t`'s
// ElementType property is not nil, it will only return true if its ElementType
// is considered the same type as `l`'s ElementType.
func (l List) Is(t Type) bool {
	return l.equals(t, false)
}

func (l List) equals(t Type, exact bool) bool {
	v, ok := t.(List)
	if !ok {
		return false
	}
	if l.ElementType == nil || v.ElementType == nil {
		// when doing exact comparisons, we can't compare types that
		// don't have element types set, so we just consider them not
		// equal
		//
		// when doing inexact comparisons, the absence of an element
		// type just means "is this a List?" We know it is, so return
		// true if and only if l has an ElementType and t doesn't. This
		// behavior only makes sense if the user is trying to see if a
		// proper type is a list, so we want to ensure that the method
		// receiver always has an element type.
		if exact {
			return false
		}
		return l.ElementType != nil
	}
	return l.ElementType.equals(v.ElementType, exact)
}

func (l List) String() string {
	return "tftypes.List[" + l.ElementType.String() + "]"
}

func (l List) private() {}

func (l List) supportedGoTypes() []string {
	return []string{"[]tftypes.Value"}
}

func valueFromList(typ Type, in interface{}) (Value, error) {
	switch value := in.(type) {
	case []Value:
		var valType Type
		for pos, v := range value {
			if err := useTypeAs(v.Type(), typ, NewAttributePath().WithElementKeyInt(int64(pos))); err != nil {
				return Value{}, err
			}
			if valType == nil {
				valType = v.Type()
			}
			if !v.Type().equals(valType, true) {
				return Value{}, fmt.Errorf("lists must only contain one type of element, saw %s and %s", valType, v.Type())
			}
		}
		return Value{
			typ:   List{ElementType: typ},
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.List; expected types are: %s", in, formattedSupportedGoTypes(List{}))
	}
}

// MarshalJSON returns a JSON representation of the full type signature of `l`,
// including its ElementType.
//
// Deprecated: this is not meant to be called by third-party code.
func (l List) MarshalJSON() ([]byte, error) {
	elementType, err := l.ElementType.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling tftypes.List's element type %T to JSON: %w", l.ElementType, err)
	}
	return []byte(`["list",` + string(elementType) + `]`), nil
}

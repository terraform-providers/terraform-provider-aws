package tftypes

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Tuple is a Terraform type representing an ordered collection of elements,
// potentially of differing types. The number of elements and their types are
// part of the type signature for the Tuple, and so two Tuples with different
// numbers or types of elements are considered to be distinct types.
type Tuple struct {
	ElementTypes []Type

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

// Equal returns true if the two Tuples are exactly equal. Unlike Is, passing
// in a Tuple with no ElementTypes will always return false.
func (tu Tuple) Equal(o Tuple) bool {
	return tu.equals(o, true)
}

// Is returns whether `t` is a Tuple type or not. If `t` is an instance of the
// Tuple type and its ElementTypes property is not nil, it will only return
// true if the ElementTypes are considered the same. To be considered the same,
// there must be the same number of ElementTypes, arranged in the same order,
// and the types in each position must be considered the same as the type in
// the same position in the other Tuple.
func (tu Tuple) Is(t Type) bool {
	return tu.equals(t, false)
}

func (tu Tuple) equals(t Type, exact bool) bool {
	v, ok := t.(Tuple)
	if !ok {
		return false
	}
	if v.ElementTypes == nil || tu.ElementTypes == nil {
		// when doing exact comparisons, we can't compare types that
		// don't have element types set, so we just consider them not
		// equal
		//
		// when doing inexact comparisons, the absence of an element
		// type just means "is this a Tuple?" We know it is, so return
		// true
		return !exact
	}
	if len(v.ElementTypes) != len(tu.ElementTypes) {
		return false
	}
	for pos, typ := range tu.ElementTypes {
		if !typ.equals(v.ElementTypes[pos], exact) {
			return false
		}
	}
	return true
}

func (tu Tuple) String() string {
	var res strings.Builder
	res.WriteString("tftypes.Tuple[")
	for pos, t := range tu.ElementTypes {
		if pos != 0 {
			res.WriteString(", ")
		}
		res.WriteString(t.String())
	}
	res.WriteString("]")
	return res.String()
}

func (tu Tuple) private() {}

func (tu Tuple) supportedGoTypes() []string {
	return []string{"[]tftypes.Value"}
}

func valueCanBeTuple(val interface{}) bool {
	switch val.(type) {
	case []Value:
		return true
	default:
		return false
	}
}

func valueFromTuple(types []Type, in interface{}) (Value, error) {
	switch value := in.(type) {
	case []Value:
		// types should only be null if the "Tuple" is actually a
		// DynamicPseudoType being created from a []Value. In which
		// case, we don't know what types it should have, or even how
		// many there will be, so let's not validate that at all
		if types != nil {
			if len(types) != len(value) {
				return Value{}, fmt.Errorf("can't create a tftypes.Value with %d elements, type %s requires %d elements", len(value), Tuple{ElementTypes: types}, len(types))
			}
			for pos, v := range value {
				typ := types[pos]
				err := useTypeAs(v.Type(), typ, NewAttributePath().WithElementKeyInt(int64(pos)))
				if err != nil {
					return Value{}, err
				}
			}
		}
		return Value{
			typ:   Tuple{ElementTypes: types},
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.Tuple; expected types are: %s", in, formattedSupportedGoTypes(Tuple{}))
	}
}

// MarshalJSON returns a JSON representation of the full type signature of
// `tu`, including the ElementTypes.
//
// Deprecated: this is not meant to be called by third-party code.
func (tu Tuple) MarshalJSON() ([]byte, error) {
	elements, err := json.Marshal(tu.ElementTypes)
	if err != nil {
		return nil, err
	}
	return []byte(`["tuple",` + string(elements) + `]`), nil
}

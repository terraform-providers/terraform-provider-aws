package tftypes

import "encoding/json"

// Tuple is a Terraform type representing an ordered collection of elements,
// potentially of differing types. The number of elements and their types are
// part of the type signature for the Tuple, and so two Tuples with different
// numbers or types of elements are considered to be distinct types.
type Tuple struct {
	ElementTypes []Type
}

// Is returns whether `t` is a Tuple type or not. If `t` is an instance of the
// Tuple type and its ElementTypes property is not nil, it will only return
// true if the ElementTypes are considered the same. To be considered the same,
// there must be the same number of ElementTypes, arranged in the same order,
// and the types in each position must be considered the same as the type in
// the same position in the other Tuple.
func (tu Tuple) Is(t Type) bool {
	v, ok := t.(Tuple)
	if !ok {
		return false
	}
	if v.ElementTypes != nil {
		if len(v.ElementTypes) != len(tu.ElementTypes) {
			return false
		}
		for pos, typ := range tu.ElementTypes {
			if !typ.Is(v.ElementTypes[pos]) {
				return false
			}
		}
	}
	return ok
}

func (tu Tuple) String() string {
	return "tftypes.Tuple"
}

func (tu Tuple) private() {}

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

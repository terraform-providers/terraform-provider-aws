package tftypes

import "fmt"

// List is a Terraform type representing an ordered collection of elements, all
// of the same type.
type List struct {
	ElementType Type
}

// Is returns whether `t` is a List type or not. If `t` is an instance of the
// List type and its ElementType property is nil, it will return true. If `t`'s
// ElementType property is not nil, it will only return true if its ElementType
// is considered the same type as `l`'s ElementType.
func (l List) Is(t Type) bool {
	v, ok := t.(List)
	if !ok {
		return false
	}
	if v.ElementType != nil {
		return l.ElementType.Is(v.ElementType)
	}
	return ok
}

func (l List) String() string {
	return "tftypes.List"
}

func (l List) private() {}

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

package schema

import (
	"go/types"
)

const (
	TypeNameResourceData = `ResourceData`
)

// IsTypeResourceData returns if the type is ResourceData from the helper/schema package
func IsTypeResourceData(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameResourceData)
	case *types.Pointer:
		return IsTypeResourceData(t.Elem())
	default:
		return false
	}
}

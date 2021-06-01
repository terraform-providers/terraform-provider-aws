package schema

import (
	"go/types"
)

const (
	TypeNameResourceDiff = `ResourceDiff`
)

// IsTypeResourceDiff returns if the type is ResourceDiff from the schema package
func IsTypeResourceDiff(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameResourceDiff)
	case *types.Pointer:
		return IsTypeResourceDiff(t.Elem())
	default:
		return false
	}
}

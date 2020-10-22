package keyvaluetags

import (
	"go/types"
)

const (
	TypeNameIgnoreConfig = `IgnoreConfig`
)

// IsTypeIgnoreConfig returns if the type is IgnoreConfig from the internal/keyvaluetags package
func IsTypeIgnoreConfig(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameKeyValueTags)
	case *types.Pointer:
		return IsTypeKeyValueTags(t.Elem())
	default:
		return false
	}
}

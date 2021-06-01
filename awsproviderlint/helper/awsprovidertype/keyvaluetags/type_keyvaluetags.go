package keyvaluetags

import (
	"go/types"
)

const (
	KeyValueTagsMethodNameChunks                 = `Chunks`
	KeyValueTagsMethodNameContainsAll            = `ContainsAll`
	KeyValueTagsMethodNameHash                   = `Hash`
	KeyValueTagsMethodNameIgnore                 = `Ignore`
	KeyValueTagsMethodNameIgnoreAws              = `IgnoreAws`
	KeyValueTagsMethodNameIgnoreConfig           = `IgnoreConfig`
	KeyValueTagsMethodNameIgnoreElasticbeanstalk = `IgnoreElasticbeanstalk`
	KeyValueTagsMethodNameIgnorePrefixes         = `IgnorePrefixes`
	KeyValueTagsMethodNameIgnoreRds              = `IgnoreRds`
	KeyValueTagsMethodNameKeys                   = `Keys`
	KeyValueTagsMethodNameMap                    = `Map`
	KeyValueTagsMethodNameMerge                  = `Merge`
	KeyValueTagsMethodNameRemoved                = `Removed`
	KeyValueTagsMethodNameUpdated                = `Updated`
	KeyValueTagsMethodNameUrlEncode              = `UrlEncode`

	TypeNameKeyValueTags = `KeyValueTags`
)

// IsTypeKeyValueTags returns if the type is KeyValueTags from the internal/keyvaluetags package
func IsTypeKeyValueTags(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameKeyValueTags)
	case *types.Pointer:
		return IsTypeKeyValueTags(t.Elem())
	default:
		return false
	}
}

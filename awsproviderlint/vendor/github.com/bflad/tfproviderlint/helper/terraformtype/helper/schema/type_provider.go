package schema

import "go/types"

const (
	ProviderFieldConfigureFunc    = `ConfigureFunc`
	ProviderFieldDataSourcesMap   = `DataSourcesMap`
	ProviderFieldMetaReset        = `MetaReset`
	ProviderFieldResourcesMap     = `ResourcesMap`
	ProviderFieldSchema           = `Schema`
	ProviderFieldTerraformVersion = `TerraformVersion`

	TypeNameProvider = `Provider`
)

// IsTypeProvider returns if the type is Provider from the schema package
func IsTypeProvider(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameProvider)
	case *types.Pointer:
		return IsTypeProvider(t.Elem())
	default:
		return false
	}
}

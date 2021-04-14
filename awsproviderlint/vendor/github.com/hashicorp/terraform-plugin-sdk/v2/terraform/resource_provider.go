package terraform

// ResourceType is a type of resource that a resource provider can manage.
type ResourceType struct {
	Name       string // Name of the resource, example "instance" (no provider prefix)
	Importable bool   // Whether this resource supports importing

	// SchemaAvailable is set if the provider supports the ProviderSchema,
	// ResourceTypeSchema and DataSourceSchema methods. Although it is
	// included on each resource type, it's actually a provider-wide setting
	// that's smuggled here only because that avoids a breaking change to
	// the plugin protocol.
	SchemaAvailable bool
}

// DataSource is a data source that a resource provider implements.
type DataSource struct {
	Name string

	// SchemaAvailable is set if the provider supports the ProviderSchema,
	// ResourceTypeSchema and DataSourceSchema methods. Although it is
	// included on each resource type, it's actually a provider-wide setting
	// that's smuggled here only because that avoids a breaking change to
	// the plugin protocol.
	SchemaAvailable bool
}

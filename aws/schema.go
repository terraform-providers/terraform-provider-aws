package aws

// resourceDiffer exposes the interface for accessing changes in a resource
// Implementations:
// * schema.ResourceData
// * schema.ResourceDiff
// FIXME: can be removed if https://github.com/hashicorp/terraform-plugin-sdk/pull/626/files is merged
type resourceDiffer interface {
	HasChange(string) bool
}

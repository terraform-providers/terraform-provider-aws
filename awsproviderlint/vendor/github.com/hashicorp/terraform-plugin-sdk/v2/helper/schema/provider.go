package schema

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const uaEnvVar = "TF_APPEND_USER_AGENT"

var ReservedProviderFields = []string{
	"alias",
	"version",
}

// StopContext returns a context safe for global use that will cancel
// when Terraform requests a stop. This function should only be called
// within a ConfigureContextFunc, passing in the request scoped context
// received in that method.
//
// Deprecated: The use of a global context is discouraged. Please use the new
// context aware CRUD methods.
func StopContext(ctx context.Context) (context.Context, bool) {
	stopContext, ok := ctx.Value(StopContextKey).(context.Context)
	return stopContext, ok
}

// Provider represents a resource provider in Terraform, and properly
// implements all of the ResourceProvider API.
//
// By defining a schema for the configuration of the provider, the
// map of supporting resources, and a configuration function, the schema
// framework takes over and handles all the provider operations for you.
//
// After defining the provider structure, it is unlikely that you'll require any
// of the methods on Provider itself.
type Provider struct {
	// Schema is the schema for the configuration of this provider. If this
	// provider has no configuration, this can be omitted.
	//
	// The keys of this map are the configuration keys, and the value is
	// the schema describing the value of the configuration.
	Schema map[string]*Schema

	// ResourcesMap is the list of available resources that this provider
	// can manage, along with their Resource structure defining their
	// own schemas and CRUD operations.
	//
	// Provider automatically handles routing operations such as Apply,
	// Diff, etc. to the proper resource.
	ResourcesMap map[string]*Resource

	// DataSourcesMap is the collection of available data sources that
	// this provider implements, with a Resource instance defining
	// the schema and Read operation of each.
	//
	// Resource instances for data sources must have a Read function
	// and must *not* implement Create, Update or Delete.
	DataSourcesMap map[string]*Resource

	// ProviderMetaSchema is the schema for the configuration of the meta
	// information for this provider. If this provider has no meta info,
	// this can be omitted. This functionality is currently experimental
	// and subject to change or break without warning; it should only be
	// used by providers that are collaborating on its use with the
	// Terraform team.
	ProviderMetaSchema map[string]*Schema

	// ConfigureFunc is a function for configuring the provider. If the
	// provider doesn't need to be configured, this can be omitted.
	//
	// Deprecated: Please use ConfigureContextFunc instead.
	ConfigureFunc ConfigureFunc

	// ConfigureContextFunc is a function for configuring the provider. If the
	// provider doesn't need to be configured, this can be omitted. This function
	// receives a context.Context that will cancel when Terraform sends a
	// cancellation signal. This function can yield Diagnostics.
	ConfigureContextFunc ConfigureContextFunc

	meta interface{}

	TerraformVersion string
}

// ConfigureFunc is the function used to configure a Provider.
//
// Deprecated: Please use ConfigureContextFunc
type ConfigureFunc func(*ResourceData) (interface{}, error)

// ConfigureContextFunc is the function used to configure a Provider.
//
// The interface{} value returned by this function is stored and passed into
// the subsequent resources as the meta parameter. This return value is
// usually used to pass along a configured API client, a configuration
// structure, etc.
type ConfigureContextFunc func(context.Context, *ResourceData) (interface{}, diag.Diagnostics)

// InternalValidate should be called to validate the structure
// of the provider.
//
// This should be called in a unit test for any provider to verify
// before release that a provider is properly configured for use with
// this library.
func (p *Provider) InternalValidate() error {
	if p == nil {
		return errors.New("provider is nil")
	}

	if p.ConfigureFunc != nil && p.ConfigureContextFunc != nil {
		return errors.New("ConfigureFunc and ConfigureContextFunc must not both be set")
	}

	var validationErrors error
	sm := schemaMap(p.Schema)
	if err := sm.InternalValidate(sm); err != nil {
		validationErrors = multierror.Append(validationErrors, err)
	}

	// Provider-specific checks
	for k := range sm {
		if isReservedProviderFieldName(k) {
			return fmt.Errorf("%s is a reserved field name for a provider", k)
		}
	}

	for k, r := range p.ResourcesMap {
		if err := r.InternalValidate(nil, true); err != nil {
			validationErrors = multierror.Append(validationErrors, fmt.Errorf("resource %s: %s", k, err))
		}
	}

	for k, r := range p.DataSourcesMap {
		if err := r.InternalValidate(nil, false); err != nil {
			validationErrors = multierror.Append(validationErrors, fmt.Errorf("data source %s: %s", k, err))
		}
	}

	return validationErrors
}

func isReservedProviderFieldName(name string) bool {
	for _, reservedName := range ReservedProviderFields {
		if name == reservedName {
			return true
		}
	}
	return false
}

// Meta returns the metadata associated with this provider that was
// returned by the Configure call. It will be nil until Configure is called.
func (p *Provider) Meta() interface{} {
	return p.meta
}

// SetMeta can be used to forcefully set the Meta object of the provider.
// Note that if Configure is called the return value will override anything
// set here.
func (p *Provider) SetMeta(v interface{}) {
	p.meta = v
}

// GetSchema returns the config schema for the main provider
// configuration, as would appear in a "provider" block in the
// configuration files.
//
// Currently not all providers support schema. Callers must therefore
// first call Resources and DataSources and ensure that at least one
// resource or data source has the SchemaAvailable flag set.
func (p *Provider) GetSchema(req *terraform.ProviderSchemaRequest) (*terraform.ProviderSchema, error) {
	resourceTypes := map[string]*configschema.Block{}
	dataSources := map[string]*configschema.Block{}

	for _, name := range req.ResourceTypes {
		if r, exists := p.ResourcesMap[name]; exists {
			resourceTypes[name] = r.CoreConfigSchema()
		}
	}
	for _, name := range req.DataSources {
		if r, exists := p.DataSourcesMap[name]; exists {
			dataSources[name] = r.CoreConfigSchema()
		}
	}

	return &terraform.ProviderSchema{
		Provider:      schemaMap(p.Schema).CoreConfigSchema(),
		ResourceTypes: resourceTypes,
		DataSources:   dataSources,
	}, nil
}

// Validate is called once at the beginning with the raw configuration
// (no interpolation done) and can return diagnostics
//
// This is called once with the provider configuration only. It may not
// be called at all if no provider configuration is given.
//
// This should not assume that any values of the configurations are valid.
// The primary use case of this call is to check that required keys are
// set.
func (p *Provider) Validate(c *terraform.ResourceConfig) diag.Diagnostics {
	if err := p.InternalValidate(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "InternalValidate",
				Detail: fmt.Sprintf("Internal validation of the provider failed! This is always a bug\n"+
					"with the provider itself, and not a user issue. Please report\n"+
					"this bug:\n\n%s", err),
			},
		}
	}

	return schemaMap(p.Schema).Validate(c)
}

// ValidateResource is called once at the beginning with the raw
// configuration (no interpolation done) and can return diagnostics.
//
// This is called once per resource.
//
// This should not assume any of the values in the resource configuration
// are valid since it is possible they have to be interpolated still.
// The primary use case of this call is to check that the required keys
// are set and that the general structure is correct.
func (p *Provider) ValidateResource(
	t string, c *terraform.ResourceConfig) diag.Diagnostics {
	r, ok := p.ResourcesMap[t]
	if !ok {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Provider doesn't support resource: %s", t),
			},
		}
	}

	return r.Validate(c)
}

// Configure configures the provider itself with the configuration
// given. This is useful for setting things like access keys.
//
// This won't be called at all if no provider configuration is given.
func (p *Provider) Configure(ctx context.Context, c *terraform.ResourceConfig) diag.Diagnostics {
	// No configuration
	if p.ConfigureFunc == nil && p.ConfigureContextFunc == nil {
		return nil
	}

	sm := schemaMap(p.Schema)

	// Get a ResourceData for this configuration. To do this, we actually
	// generate an intermediary "diff" although that is never exposed.
	diff, err := sm.Diff(ctx, nil, c, nil, p.meta, true)
	if err != nil {
		return diag.FromErr(err)
	}

	data, err := sm.Data(nil, diff)
	if err != nil {
		return diag.FromErr(err)
	}

	if p.ConfigureFunc != nil {
		meta, err := p.ConfigureFunc(data)
		if err != nil {
			return diag.FromErr(err)
		}
		p.meta = meta
	}
	if p.ConfigureContextFunc != nil {
		meta, diags := p.ConfigureContextFunc(ctx, data)
		if diags.HasError() {
			return diags
		}
		p.meta = meta
	}

	return nil
}

// Resources returns all the available resource types that this provider
// knows how to manage.
func (p *Provider) Resources() []terraform.ResourceType {
	keys := make([]string, 0, len(p.ResourcesMap))
	for k := range p.ResourcesMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]terraform.ResourceType, 0, len(keys))
	for _, k := range keys {
		resource := p.ResourcesMap[k]

		// This isn't really possible (it'd fail InternalValidate), but
		// we do it anyways to avoid a panic.
		if resource == nil {
			resource = &Resource{}
		}

		result = append(result, terraform.ResourceType{
			Name:       k,
			Importable: resource.Importer != nil,

			// Indicates that a provider is compiled against a new enough
			// version of core to support the GetSchema method.
			SchemaAvailable: true,
		})
	}

	return result
}

// ImportState requests that the given resource be imported.
//
// The returned InstanceState only requires ID be set. Importing
// will always call Refresh after the state to complete it.
//
// IMPORTANT: InstanceState doesn't have the resource type attached
// to it. A type must be specified on the state via the Ephemeral
// field on the state.
//
// This function can return multiple states. Normally, an import
// will map 1:1 to a physical resource. However, some resources map
// to multiple. For example, an AWS security group may contain many rules.
// Each rule is represented by a separate resource in Terraform,
// therefore multiple states are returned.
func (p *Provider) ImportState(
	ctx context.Context,
	info *terraform.InstanceInfo,
	id string) ([]*terraform.InstanceState, error) {
	// Find the resource
	r, ok := p.ResourcesMap[info.Type]
	if !ok {
		return nil, fmt.Errorf("unknown resource type: %s", info.Type)
	}

	// If it doesn't support import, error
	if r.Importer == nil {
		return nil, fmt.Errorf("resource %s doesn't support import", info.Type)
	}

	// Create the data
	data := r.Data(nil)
	data.SetId(id)
	data.SetType(info.Type)

	// Call the import function
	results := []*ResourceData{data}
	if r.Importer.State != nil || r.Importer.StateContext != nil {
		var err error
		if r.Importer.StateContext != nil {
			results, err = r.Importer.StateContext(ctx, data, p.meta)
		} else {
			results, err = r.Importer.State(data, p.meta)
		}
		if err != nil {
			return nil, err
		}
	}

	// Convert the results to InstanceState values and return it
	states := make([]*terraform.InstanceState, len(results))
	for i, r := range results {
		states[i] = r.State()
	}

	// Verify that all are non-nil. If there are any nil the error
	// isn't obvious so we circumvent that with a friendlier error.
	for _, s := range states {
		if s == nil {
			return nil, fmt.Errorf(
				"nil entry in ImportState results. This is always a bug with\n" +
					"the resource that is being imported. Please report this as\n" +
					"a bug to Terraform.")
		}
	}

	return states, nil
}

// ValidateDataSource is called once at the beginning with the raw
// configuration (no interpolation done) and can return diagnostics.
//
// This is called once per data source instance.
//
// This should not assume any of the values in the resource configuration
// are valid since it is possible they have to be interpolated still.
// The primary use case of this call is to check that the required keys
// are set and that the general structure is correct.
func (p *Provider) ValidateDataSource(
	t string, c *terraform.ResourceConfig) diag.Diagnostics {
	r, ok := p.DataSourcesMap[t]
	if !ok {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Provider doesn't support data source: %s", t),
			},
		}
	}

	return r.Validate(c)
}

// DataSources returns all of the available data sources that this
// provider implements.
func (p *Provider) DataSources() []terraform.DataSource {
	keys := make([]string, 0, len(p.DataSourcesMap))
	for k := range p.DataSourcesMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]terraform.DataSource, 0, len(keys))
	for _, k := range keys {
		result = append(result, terraform.DataSource{
			Name: k,

			// Indicates that a provider is compiled against a new enough
			// version of core to support the GetSchema method.
			SchemaAvailable: true,
		})
	}

	return result
}

// UserAgent returns a string suitable for use in the User-Agent header of
// requests generated by the provider. The generated string contains the
// version of Terraform, the Plugin SDK, and the provider used to generate the
// request. `name` should be the hyphen-separated reporting name of the
// provider, and `version` should be the version of the provider.
//
// If TF_APPEND_USER_AGENT is set, its value will be appended to the returned
// string.
func (p *Provider) UserAgent(name, version string) string {
	ua := fmt.Sprintf("Terraform/%s (+https://www.terraform.io) Terraform-Plugin-SDK/%s", p.TerraformVersion, meta.SDKVersionString())
	if name != "" {
		ua += " " + name
		if version != "" {
			ua += "/" + version
		}
	}

	if add := os.Getenv(uaEnvVar); add != "" {
		add = strings.TrimSpace(add)
		if len(add) > 0 {
			ua += " " + add
			log.Printf("[DEBUG] Using modified User-Agent: %s", ua)
		}
	}

	return ua
}

// GRPCProvider returns a gRPC server, for use with terraform-plugin-mux.
func (p *Provider) GRPCProvider() tfprotov5.ProviderServer {
	return NewGRPCProviderServer(p)
}

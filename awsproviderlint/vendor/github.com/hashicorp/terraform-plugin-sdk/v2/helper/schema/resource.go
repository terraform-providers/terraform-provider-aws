package schema

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var ReservedDataSourceFields = []string{
	"connection",
	"count",
	"depends_on",
	"lifecycle",
	"provider",
	"provisioner",
}

var ReservedResourceFields = []string{
	"connection",
	"count",
	"depends_on",
	"lifecycle",
	"provider",
	"provisioner",
}

// Resource represents a thing in Terraform that has a set of configurable
// attributes and a lifecycle (create, read, update, delete).
//
// The Resource schema is an abstraction that allows provider writers to
// worry only about CRUD operations while off-loading validation, diff
// generation, etc. to this higher level library.
//
// In spite of the name, this struct is not used only for terraform resources,
// but also for data sources. In the case of data sources, the Create,
// Update and Delete functions must not be provided.
type Resource struct {
	// Schema is the schema for the configuration of this resource.
	//
	// The keys of this map are the configuration keys, and the values
	// describe the schema of the configuration value.
	//
	// The schema is used to represent both configurable data as well
	// as data that might be computed in the process of creating this
	// resource.
	Schema map[string]*Schema

	// SchemaVersion is the version number for this resource's Schema
	// definition. The current SchemaVersion stored in the state for each
	// resource. Provider authors can increment this version number
	// when Schema semantics change. If the State's SchemaVersion is less than
	// the current SchemaVersion, the InstanceState is yielded to the
	// MigrateState callback, where the provider can make whatever changes it
	// needs to update the state to be compatible to the latest version of the
	// Schema.
	//
	// When unset, SchemaVersion defaults to 0, so provider authors can start
	// their Versioning at any integer >= 1
	SchemaVersion int

	// MigrateState is responsible for updating an InstanceState with an old
	// version to the format expected by the current version of the Schema.
	//
	// It is called during Refresh if the State's stored SchemaVersion is less
	// than the current SchemaVersion of the Resource.
	//
	// The function is yielded the state's stored SchemaVersion and a pointer to
	// the InstanceState that needs updating, as well as the configured
	// provider's configured meta interface{}, in case the migration process
	// needs to make any remote API calls.
	//
	// Deprecated: MigrateState is deprecated and any new changes to a resource's schema
	// should be handled by StateUpgraders. Existing MigrateState implementations
	// should remain for compatibility with existing state. MigrateState will
	// still be called if the stored SchemaVersion is less than the
	// first version of the StateUpgraders.
	MigrateState StateMigrateFunc

	// StateUpgraders contains the functions responsible for upgrading an
	// existing state with an old schema version to a newer schema. It is
	// called specifically by Terraform when the stored schema version is less
	// than the current SchemaVersion of the Resource.
	//
	// StateUpgraders map specific schema versions to a StateUpgrader
	// function. The registered versions are expected to be ordered,
	// consecutive values. The initial value may be greater than 0 to account
	// for legacy schemas that weren't recorded and can be handled by
	// MigrateState.
	StateUpgraders []StateUpgrader

	// The functions below are the CRUD operations for this resource.
	//
	// Deprecated: Please use the context aware equivalents instead. Only one of
	// the operations or context aware equivalent can be set, not both.
	Create CreateFunc
	Read   ReadFunc
	Update UpdateFunc
	Delete DeleteFunc

	// Exists is a function that is called to check if a resource still
	// exists. If this returns false, then this will affect the diff
	// accordingly. If this function isn't set, it will not be called. You
	// can also signal existence in the Read method by calling d.SetId("")
	// if the Resource is no longer present and should be removed from state.
	// The *ResourceData passed to Exists should _not_ be modified.
	//
	// Deprecated: ReadContext should be able to encapsulate the logic of Exists
	Exists ExistsFunc

	// The functions below are the CRUD operations for this resource.
	//
	// The only optional operation is Update. If Update is not
	// implemented, then updates will not be supported for this resource.
	//
	// The ResourceData parameter in the functions below are used to
	// query configuration and changes for the resource as well as to set
	// the ID, computed data, etc.
	//
	// The interface{} parameter is the result of the ConfigureFunc in
	// the provider for this resource. If the provider does not define
	// a ConfigureFunc, this will be nil. This parameter should be used
	// to store API clients, configuration structures, etc.
	//
	// These functions are passed a context configured to timeout with whatever
	// was set as the timeout for this operation. Useful for forwarding on to
	// backend SDK's that accept context. The context will also cancel if
	// Terraform sends a cancellation signal.
	//
	// These functions return diagnostics, allowing developers to build
	// a list of warnings and errors to be presented to the Terraform user.
	// The AttributePath of those diagnostics should be built within these
	// functions, please consult go-cty documentation for building a cty.Path
	CreateContext CreateContextFunc
	ReadContext   ReadContextFunc
	UpdateContext UpdateContextFunc
	DeleteContext DeleteContextFunc

	// CustomizeDiff is a custom function for working with the diff that
	// Terraform has created for this resource - it can be used to customize the
	// diff that has been created, diff values not controlled by configuration,
	// or even veto the diff altogether and abort the plan. It is passed a
	// *ResourceDiff, a structure similar to ResourceData but lacking most write
	// functions like Set, while introducing new functions that work with the
	// diff such as SetNew, SetNewComputed, and ForceNew.
	//
	// The phases Terraform runs this in, and the state available via functions
	// like Get and GetChange, are as follows:
	//
	//  * New resource: One run with no state
	//  * Existing resource: One run with state
	//   * Existing resource, forced new: One run with state (before ForceNew),
	//     then one run without state (as if new resource)
	//  * Tainted resource: No runs (custom diff logic is skipped)
	//  * Destroy: No runs (standard diff logic is skipped on destroy diffs)
	//
	// This function needs to be resilient to support all scenarios.
	//
	// For the most part, only computed fields can be customized by this
	// function.
	//
	// This function is only allowed on regular resources (not data sources).
	CustomizeDiff CustomizeDiffFunc

	// Importer is the ResourceImporter implementation for this resource.
	// If this is nil, then this resource does not support importing. If
	// this is non-nil, then it supports importing and ResourceImporter
	// must be validated. The validity of ResourceImporter is verified
	// by InternalValidate on Resource.
	Importer *ResourceImporter

	// If non-empty, this string is emitted as a warning during Validate.
	DeprecationMessage string

	// Timeouts allow users to specify specific time durations in which an
	// operation should time out, to allow them to extend an action to suit their
	// usage. For example, a user may specify a large Creation timeout for their
	// AWS RDS Instance due to it's size, or restoring from a snapshot.
	// Resource implementors must enable Timeout support by adding the allowed
	// actions (Create, Read, Update, Delete, Default) to the Resource struct, and
	// accessing them in the matching methods.
	Timeouts *ResourceTimeout

	// Description is used as the description for docs, the language server and
	// other user facing usage. It can be plain-text or markdown depending on the
	// global DescriptionKind setting.
	Description string

	// UseJSONNumber should be set when state upgraders will expect
	// json.Numbers instead of float64s for numbers. This is added as a
	// toggle for backwards compatibility for type assertions, but should
	// be used in all new resources to avoid bugs with sufficiently large
	// user input.
	//
	// See github.com/hashicorp/terraform-plugin-sdk/issues/655 for more
	// details.
	UseJSONNumber bool
}

// ShimInstanceStateFromValue converts a cty.Value to a
// terraform.InstanceState.
func (r *Resource) ShimInstanceStateFromValue(state cty.Value) (*terraform.InstanceState, error) {
	// Get the raw shimmed value. While this is correct, the set hashes don't
	// match those from the Schema.
	s := terraform.NewInstanceStateShimmedFromValue(state, r.SchemaVersion)

	// We now rebuild the state through the ResourceData, so that the set indexes
	// match what helper/schema expects.
	data, err := schemaMap(r.Schema).Data(s, nil)
	if err != nil {
		return nil, err
	}

	s = data.State()
	if s == nil {
		s = &terraform.InstanceState{}
	}
	return s, nil
}

// The following function types are of the legacy CRUD operations.
//
// Deprecated: Please use the context aware equivalents instead.
type CreateFunc func(*ResourceData, interface{}) error
type ReadFunc func(*ResourceData, interface{}) error
type UpdateFunc func(*ResourceData, interface{}) error
type DeleteFunc func(*ResourceData, interface{}) error
type ExistsFunc func(*ResourceData, interface{}) (bool, error)

// See Resource documentation.
type CreateContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics

// See Resource documentation.
type ReadContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics

// See Resource documentation.
type UpdateContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics

// See Resource documentation.
type DeleteContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics

// See Resource documentation.
type StateMigrateFunc func(
	int, *terraform.InstanceState, interface{}) (*terraform.InstanceState, error)

type StateUpgrader struct {
	// Version is the version schema that this Upgrader will handle, converting
	// it to Version+1.
	Version int

	// Type describes the schema that this function can upgrade. Type is
	// required to decode the schema if the state was stored in a legacy
	// flatmap format.
	Type cty.Type

	// Upgrade takes the JSON encoded state and the provider meta value, and
	// upgrades the state one single schema version. The provided state is
	// deocded into the default json types using a map[string]interface{}. It
	// is up to the StateUpgradeFunc to ensure that the returned value can be
	// encoded using the new schema.
	Upgrade StateUpgradeFunc
}

// See StateUpgrader
type StateUpgradeFunc func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error)

// See Resource documentation.
type CustomizeDiffFunc func(context.Context, *ResourceDiff, interface{}) error

func (r *Resource) create(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
	if r.Create != nil {
		if err := r.Create(d, meta); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, d.Timeout(TimeoutCreate))
	defer cancel()
	return r.CreateContext(ctx, d, meta)
}

func (r *Resource) read(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
	if r.Read != nil {
		if err := r.Read(d, meta); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, d.Timeout(TimeoutRead))
	defer cancel()
	return r.ReadContext(ctx, d, meta)
}

func (r *Resource) update(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
	if r.Update != nil {
		if err := r.Update(d, meta); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, d.Timeout(TimeoutUpdate))
	defer cancel()
	return r.UpdateContext(ctx, d, meta)
}

func (r *Resource) delete(ctx context.Context, d *ResourceData, meta interface{}) diag.Diagnostics {
	if r.Delete != nil {
		if err := r.Delete(d, meta); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, d.Timeout(TimeoutDelete))
	defer cancel()
	return r.DeleteContext(ctx, d, meta)
}

// Apply creates, updates, and/or deletes a resource.
func (r *Resource) Apply(
	ctx context.Context,
	s *terraform.InstanceState,
	d *terraform.InstanceDiff,
	meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
	data, err := schemaMap(r.Schema).Data(s, d)
	if err != nil {
		return s, diag.FromErr(err)
	}

	if s != nil && data != nil {
		data.providerMeta = s.ProviderMeta
	}

	// Instance Diff shoould have the timeout info, need to copy it over to the
	// ResourceData meta
	rt := ResourceTimeout{}
	if _, ok := d.Meta[TimeoutKey]; ok {
		if err := rt.DiffDecode(d); err != nil {
			log.Printf("[ERR] Error decoding ResourceTimeout: %s", err)
		}
	} else if s != nil {
		if _, ok := s.Meta[TimeoutKey]; ok {
			if err := rt.StateDecode(s); err != nil {
				log.Printf("[ERR] Error decoding ResourceTimeout: %s", err)
			}
		}
	} else {
		log.Printf("[DEBUG] No meta timeoutkey found in Apply()")
	}
	data.timeouts = &rt

	if s == nil {
		// The Terraform API dictates that this should never happen, but
		// it doesn't hurt to be safe in this case.
		s = new(terraform.InstanceState)
	}

	var diags diag.Diagnostics

	if d.Destroy || d.RequiresNew() {
		if s.ID != "" {
			// Destroy the resource since it is created
			diags = append(diags, r.delete(ctx, data, meta)...)
			if diags.HasError() {
				return r.recordCurrentSchemaVersion(data.State()), diags
			}

			// Make sure the ID is gone.
			data.SetId("")
		}

		// If we're only destroying, and not creating, then return
		// now since we're done!
		if !d.RequiresNew() {
			return nil, diags
		}

		// Reset the data to be stateless since we just destroyed
		data, err = schemaMap(r.Schema).Data(nil, d)
		if err != nil {
			return nil, append(diags, diag.FromErr(err)...)
		}

		// data was reset, need to re-apply the parsed timeouts
		data.timeouts = &rt
	}

	if data.Id() == "" {
		// We're creating, it is a new resource.
		data.MarkNewResource()
		diags = append(diags, r.create(ctx, data, meta)...)
	} else {
		if !r.updateFuncSet() {
			return s, append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "doesn't support update",
			})
		}
		diags = append(diags, r.update(ctx, data, meta)...)
	}

	return r.recordCurrentSchemaVersion(data.State()), diags
}

// Diff returns a diff of this resource.
func (r *Resource) Diff(
	ctx context.Context,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.InstanceDiff, error) {

	t := &ResourceTimeout{}
	err := t.ConfigDecode(r, c)

	if err != nil {
		return nil, fmt.Errorf("[ERR] Error decoding timeout: %s", err)
	}

	instanceDiff, err := schemaMap(r.Schema).Diff(ctx, s, c, r.CustomizeDiff, meta, true)
	if err != nil {
		return instanceDiff, err
	}

	if instanceDiff != nil {
		if err := t.DiffEncode(instanceDiff); err != nil {
			log.Printf("[ERR] Error encoding timeout to instance diff: %s", err)
		}
	} else {
		log.Printf("[DEBUG] Instance Diff is nil in Diff()")
	}

	return instanceDiff, err
}

func (r *Resource) SimpleDiff(
	ctx context.Context,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.InstanceDiff, error) {

	instanceDiff, err := schemaMap(r.Schema).Diff(ctx, s, c, r.CustomizeDiff, meta, false)
	if err != nil {
		return instanceDiff, err
	}

	if instanceDiff == nil {
		instanceDiff = terraform.NewInstanceDiff()
	}

	// Make sure the old value is set in each of the instance diffs.
	// This was done by the RequiresNew logic in the full legacy Diff.
	for k, attr := range instanceDiff.Attributes {
		if attr == nil {
			continue
		}
		if s != nil {
			attr.Old = s.Attributes[k]
		}
	}

	return instanceDiff, nil
}

// Validate validates the resource configuration against the schema.
func (r *Resource) Validate(c *terraform.ResourceConfig) diag.Diagnostics {
	diags := schemaMap(r.Schema).Validate(c)

	if r.DeprecationMessage != "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Deprecated Resource",
			Detail:   r.DeprecationMessage,
		})
	}

	return diags
}

// ReadDataApply loads the data for a data source, given a diff that
// describes the configuration arguments and desired computed attributes.
func (r *Resource) ReadDataApply(
	ctx context.Context,
	d *terraform.InstanceDiff,
	meta interface{},
) (*terraform.InstanceState, diag.Diagnostics) {
	// Data sources are always built completely from scratch
	// on each read, so the source state is always nil.
	data, err := schemaMap(r.Schema).Data(nil, d)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	diags := r.read(ctx, data, meta)
	state := data.State()
	if state != nil && state.ID == "" {
		// Data sources can set an ID if they want, but they aren't
		// required to; we'll provide a placeholder if they don't,
		// to preserve the invariant that all resources have non-empty
		// ids.
		state.ID = "-"
	}

	return r.recordCurrentSchemaVersion(state), diags
}

// RefreshWithoutUpgrade reads the instance state, but does not call
// MigrateState or the StateUpgraders, since those are now invoked in a
// separate API call.
// RefreshWithoutUpgrade is part of the new plugin shims.
func (r *Resource) RefreshWithoutUpgrade(
	ctx context.Context,
	s *terraform.InstanceState,
	meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
	// If the ID is already somehow blank, it doesn't exist
	if s.ID == "" {
		return nil, nil
	}

	rt := ResourceTimeout{}
	if _, ok := s.Meta[TimeoutKey]; ok {
		if err := rt.StateDecode(s); err != nil {
			log.Printf("[ERR] Error decoding ResourceTimeout: %s", err)
		}
	}

	if r.Exists != nil {
		// Make a copy of data so that if it is modified it doesn't
		// affect our Read later.
		data, err := schemaMap(r.Schema).Data(s, nil)
		if err != nil {
			return s, diag.FromErr(err)
		}
		data.timeouts = &rt

		if s != nil {
			data.providerMeta = s.ProviderMeta
		}

		exists, err := r.Exists(data, meta)
		if err != nil {
			return s, diag.FromErr(err)
		}

		if !exists {
			return nil, nil
		}
	}

	data, err := schemaMap(r.Schema).Data(s, nil)
	if err != nil {
		return s, diag.FromErr(err)
	}
	data.timeouts = &rt

	if s != nil {
		data.providerMeta = s.ProviderMeta
	}

	diags := r.read(ctx, data, meta)
	state := data.State()
	if state != nil && state.ID == "" {
		state = nil
	}

	return r.recordCurrentSchemaVersion(state), diags
}

func (r *Resource) createFuncSet() bool {
	return (r.Create != nil || r.CreateContext != nil)
}

func (r *Resource) readFuncSet() bool {
	return (r.Read != nil || r.ReadContext != nil)
}

func (r *Resource) updateFuncSet() bool {
	return (r.Update != nil || r.UpdateContext != nil)
}

func (r *Resource) deleteFuncSet() bool {
	return (r.Delete != nil || r.DeleteContext != nil)
}

// InternalValidate should be called to validate the structure
// of the resource.
//
// This should be called in a unit test for any resource to verify
// before release that a resource is properly configured for use with
// this library.
//
// Provider.InternalValidate() will automatically call this for all of
// the resources it manages, so you don't need to call this manually if it
// is part of a Provider.
func (r *Resource) InternalValidate(topSchemaMap schemaMap, writable bool) error {
	if r == nil {
		return errors.New("resource is nil")
	}

	if !writable {
		if r.createFuncSet() || r.updateFuncSet() || r.deleteFuncSet() {
			return fmt.Errorf("must not implement Create, Update or Delete")
		}

		// CustomizeDiff cannot be defined for read-only resources
		if r.CustomizeDiff != nil {
			return fmt.Errorf("cannot implement CustomizeDiff")
		}
	}

	tsm := topSchemaMap

	if r.isTopLevel() && writable {
		// All non-Computed attributes must be ForceNew if Update is not defined
		if !r.updateFuncSet() {
			nonForceNewAttrs := make([]string, 0)
			for k, v := range r.Schema {
				if !v.ForceNew && !v.Computed {
					nonForceNewAttrs = append(nonForceNewAttrs, k)
				}
			}
			if len(nonForceNewAttrs) > 0 {
				return fmt.Errorf(
					"No Update defined, must set ForceNew on: %#v", nonForceNewAttrs)
			}
		} else {
			nonUpdateableAttrs := make([]string, 0)
			for k, v := range r.Schema {
				if v.ForceNew || v.Computed && !v.Optional {
					nonUpdateableAttrs = append(nonUpdateableAttrs, k)
				}
			}
			updateableAttrs := len(r.Schema) - len(nonUpdateableAttrs)
			if updateableAttrs == 0 {
				return fmt.Errorf(
					"All fields are ForceNew or Computed w/out Optional, Update is superfluous")
			}
		}

		tsm = schemaMap(r.Schema)

		// Destroy, and Read are required
		if !r.readFuncSet() {
			return fmt.Errorf("Read must be implemented")
		}
		if !r.deleteFuncSet() {
			return fmt.Errorf("Delete must be implemented")
		}

		// If we have an importer, we need to verify the importer.
		if r.Importer != nil {
			if err := r.Importer.InternalValidate(); err != nil {
				return err
			}
		}

		if f, ok := tsm["id"]; ok {
			// if there is an explicit ID, validate it...
			err := validateResourceID(f)
			if err != nil {
				return err
			}
		}

		for k, f := range tsm {
			if isReservedResourceFieldName(k, f) {
				return fmt.Errorf("%s is a reserved field name", k)
			}
		}
	}

	lastVersion := -1
	for _, u := range r.StateUpgraders {
		if lastVersion >= 0 && u.Version-lastVersion > 1 {
			return fmt.Errorf("missing schema version between %d and %d", lastVersion, u.Version)
		}

		if u.Version >= r.SchemaVersion {
			return fmt.Errorf("StateUpgrader version %d is >= current version %d", u.Version, r.SchemaVersion)
		}

		if !u.Type.IsObjectType() {
			return fmt.Errorf("StateUpgrader %d type is not cty.Object", u.Version)
		}

		if u.Upgrade == nil {
			return fmt.Errorf("StateUpgrader %d missing StateUpgradeFunc", u.Version)
		}

		lastVersion = u.Version
	}

	if lastVersion >= 0 && lastVersion != r.SchemaVersion-1 {
		return fmt.Errorf("missing StateUpgrader between %d and %d", lastVersion, r.SchemaVersion)
	}

	// Data source
	if r.isTopLevel() && !writable {
		tsm = schemaMap(r.Schema)
		for k := range tsm {
			if isReservedDataSourceFieldName(k) {
				return fmt.Errorf("%s is a reserved field name", k)
			}
		}
	}

	// check context funcs are not set alongside their nonctx counterparts
	if r.CreateContext != nil && r.Create != nil {
		return fmt.Errorf("CreateContext and Create should not both be set")
	}
	if r.ReadContext != nil && r.Read != nil {
		return fmt.Errorf("ReadContext and Read should not both be set")
	}
	if r.UpdateContext != nil && r.Update != nil {
		return fmt.Errorf("UpdateContext and Update should not both be set")
	}
	if r.DeleteContext != nil && r.Delete != nil {
		return fmt.Errorf("DeleteContext and Delete should not both be set")
	}

	return schemaMap(r.Schema).InternalValidate(tsm)
}

func isReservedDataSourceFieldName(name string) bool {
	for _, reservedName := range ReservedDataSourceFields {
		if name == reservedName {
			return true
		}
	}
	return false
}

func validateResourceID(s *Schema) error {
	if s.Type != TypeString {
		return fmt.Errorf(`the "id" attribute must be of TypeString`)
	}

	if s.Required {
		return fmt.Errorf(`the "id" attribute cannot be marked Required`)
	}

	// ID should at least be computed. If unspecified it will be set to Computed and Optional,
	// but Optional is unnecessary if undesired.
	if s.Computed != true {
		return fmt.Errorf(`the "id" attribute must be marked Computed`)
	}
	return nil
}

func isReservedResourceFieldName(name string, s *Schema) bool {
	for _, reservedName := range ReservedResourceFields {
		if name == reservedName {
			return true
		}
	}

	return false
}

// Data returns a ResourceData struct for this Resource. Each return value
// is a separate copy and can be safely modified differently.
//
// The data returned from this function has no actual affect on the Resource
// itself (including the state given to this function).
//
// This function is useful for unit tests and ResourceImporter functions.
func (r *Resource) Data(s *terraform.InstanceState) *ResourceData {
	result, err := schemaMap(r.Schema).Data(s, nil)
	if err != nil {
		// At the time of writing, this isn't possible (Data never returns
		// non-nil errors). We panic to find this in the future if we have to.
		// I don't see a reason for Data to ever return an error.
		panic(err)
	}

	// load the Resource timeouts
	result.timeouts = r.Timeouts
	if result.timeouts == nil {
		result.timeouts = &ResourceTimeout{}
	}

	// Set the schema version to latest by default
	result.meta = map[string]interface{}{
		"schema_version": strconv.Itoa(r.SchemaVersion),
	}

	return result
}

// TestResourceData Yields a ResourceData filled with this resource's schema for use in unit testing
//
// TODO: May be able to be removed with the above ResourceData function.
func (r *Resource) TestResourceData() *ResourceData {
	return &ResourceData{
		schema: r.Schema,
	}
}

// Returns true if the resource is "top level" i.e. not a sub-resource.
func (r *Resource) isTopLevel() bool {
	// TODO: This is a heuristic; replace with a definitive attribute?
	return (r.createFuncSet() || r.readFuncSet())
}

func (r *Resource) recordCurrentSchemaVersion(
	state *terraform.InstanceState) *terraform.InstanceState {
	if state != nil && r.SchemaVersion > 0 {
		if state.Meta == nil {
			state.Meta = make(map[string]interface{})
		}
		state.Meta["schema_version"] = strconv.Itoa(r.SchemaVersion)
	}
	return state
}

// Noop is a convenience implementation of resource function which takes
// no action and returns no error.
func Noop(*ResourceData, interface{}) error {
	return nil
}

// NoopContext is a convenience implementation of context aware resource function which takes
// no action and returns no error.
func NoopContext(context.Context, *ResourceData, interface{}) diag.Diagnostics {
	return nil
}

// RemoveFromState is a convenience implementation of a resource function
// which sets the resource ID to empty string (to remove it from state)
// and returns no error.
func RemoveFromState(d *ResourceData, _ interface{}) error {
	d.SetId("")
	return nil
}

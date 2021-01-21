package tfprotov5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

// ResourceServer is an interface containing the methods a resource
// implementation needs to fill.
type ResourceServer interface {
	// ValidateResourceTypeConfig is called when Terraform is checking that
	// a resource's configuration is valid. It is guaranteed to have types
	// conforming to your schema. This is your opportunity to do custom or
	// advanced validation prior to a plan being generated.
	ValidateResourceTypeConfig(context.Context, *ValidateResourceTypeConfigRequest) (*ValidateResourceTypeConfigResponse, error)

	// UpgradeResourceState is called when Terraform has encountered a
	// resource with a state in a schema that doesn't match the schema's
	// current version. It is the provider's responsibility to modify the
	// state to upgrade it to the latest state schema.
	UpgradeResourceState(context.Context, *UpgradeResourceStateRequest) (*UpgradeResourceStateResponse, error)

	// ReadResource is called when Terraform is refreshing a resource's
	// state.
	ReadResource(context.Context, *ReadResourceRequest) (*ReadResourceResponse, error)

	// PlanResourceChange is called when Terraform is attempting to
	// calculate a plan for a resource. Terraform will suggest a proposed
	// new state, which the provider can modify or return unmodified to
	// influence Terraform's plan.
	PlanResourceChange(context.Context, *PlanResourceChangeRequest) (*PlanResourceChangeResponse, error)

	// ApplyResourceChange is called when Terraform has detected a diff
	// between the resource's state and the user's config, and the user has
	// approved a planned change. The provider is to apply the changes
	// contained in the plan, and return the resulting state.
	ApplyResourceChange(context.Context, *ApplyResourceChangeRequest) (*ApplyResourceChangeResponse, error)

	// ImportResourceState is called when a user has requested Terraform
	// import a resource. The provider should fetch the information
	// specified by the passed ID and return it as one or more resource
	// states for Terraform to assume control of.
	ImportResourceState(context.Context, *ImportResourceStateRequest) (*ImportResourceStateResponse, error)
}

// ValidateResourceTypeConfigRequest is the request Terraform sends when it
// wants to validate a resource's configuration.
type ValidateResourceTypeConfigRequest struct {
	// TypeName is the type of resource Terraform is validating.
	TypeName string

	// Config is the configuration the user supplied for that resource. See
	// the documentation on `DynamicValue` for more information about
	// safely accessing the configuration.
	//
	// The configuration is represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time. Any attributes not directly
	// set in the configuration will be null.
	Config *DynamicValue
}

// ValidateResourceTypeConfigResponse is the response from the provider about
// the validity of a resource's configuration.
type ValidateResourceTypeConfigResponse struct {
	// Diagnostics report errors or warnings related to the given
	// configuration. Returning an empty slice indicates a successful
	// validation with no warnings or errors generated.
	Diagnostics []*Diagnostic
}

// UpgradeResourceStateRequest is the request Terraform sends when it needs a
// provider to upgrade the state of a given resource.
type UpgradeResourceStateRequest struct {
	// TypeName is the type of resource that Terraform needs to upgrade the
	// state for.
	TypeName string

	// Version is the version of the state the resource currently has.
	Version int64

	// RawState is the state as Terraform sees it right now. See the
	// documentation for `RawState` for information on how to work with the
	// data it contains.
	RawState *RawState
}

// UpgradeResourceStateResponse is the response from the provider containing
// the upgraded state for the given resource.
type UpgradeResourceStateResponse struct {
	// UpgradedState is the upgraded state for the resource, represented as
	// a `DynamicValue`. See the documentation on `DynamicValue` for
	// information about safely creating the `DynamicValue`.
	//
	// The state should be represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	UpgradedState *DynamicValue

	// Diagnostics report errors or warnings related to upgrading the
	// state of the requested resource. Returning an empty slice indicates
	// a successful validation with no warnings or errors generated.
	Diagnostics []*Diagnostic
}

// ReadResourceRequest is the request Terraform sends when it wants to get the
// latest state for a resource.
type ReadResourceRequest struct {
	// TypeName is the type of resource Terraform is requesting an upated
	// state for.
	TypeName string

	// CurrentState is the current state of the resource as far as
	// Terraform knows, represented as a `DynamicValue`. See the
	// documentation for `DynamicValue` for information about safely
	// accessing the state.
	//
	// The state is represented as a tftypes.Object, with each attribute
	// and nested block getting its own key and value.
	CurrentState *DynamicValue

	// Private is any provider-defined private state stored with the
	// resource. It is used for keeping state with the resource that is not
	// meant to be included when calculating diffs.
	Private []byte

	// ProviderMeta supplies the provider metadata configuration for the
	// module this resource is in. Module-specific provider metadata is an
	// advanced feature and usage of it should be coordinated with the
	// Terraform Core team by raising an issue at
	// https://github.com/hashicorp/terraform/issues/new/choose. See the
	// documentation on `DynamicValue` for information about safely
	// accessing the configuration.
	//
	// The configuration is represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	//
	// This configuration will have known values for all fields.
	ProviderMeta *DynamicValue
}

// ReadResourceResponse is the response from the provider about the current
// state of the requested resource.
type ReadResourceResponse struct {
	// NewState is the current state of the resource according to the
	// provider, represented as a `DynamicValue`. See the documentation for
	// `DynamicValue` for information about safely creating the
	// `DynamicValue`.
	//
	// The state should be represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	NewState *DynamicValue

	// Diagnostics report errors or warnings related to retrieving the
	// current state of the requested resource. Returning an empty slice
	// indicates a successful validation with no warnings or errors
	// generated.
	Diagnostics []*Diagnostic

	// Private should be set to any state that the provider would like sent
	// with requests for this resource. This state will be associated with
	// the resource, but will not be considered when calculating diffs.
	Private []byte
}

// PlanResourceChangeRequest is the request Terraform sends when it is
// generating a plan for a resource and wants the provider's input on what the
// planned state should be.
type PlanResourceChangeRequest struct {
	// TypeName is the type of resource Terraform is generating a plan for.
	TypeName string

	// PriorState is the state of the resource before the plan is applied,
	// represented as a `DynamicValue`. See the documentation for
	// `DynamicValue` for information about safely accessing the state.
	//
	// The state is represented as a tftypes.Object, with each attribute
	// and nested block getting its own key and value.
	PriorState *DynamicValue

	// ProposedNewState is the state that Terraform is proposing for the
	// resource, with the changes in the configuration applied, represented
	// as a `DynamicValue`. See the documentation for `DynamicValue` for
	// information about safely accessing the state.
	//
	// The ProposedNewState merges any non-null values in the configuration
	// with any computed attributes in PriorState as a utility to help
	// providers avoid needing to implement such merging functionality
	// themselves.
	//
	// The state is represented as a tftypes.Object, with each attribute
	// and nested block getting its own key and value.
	//
	// The ProposedNewState will be null when planning a delete operation.
	ProposedNewState *DynamicValue

	// Config is the configuration the user supplied for the resource. See
	// the documentation on `DynamicValue` for more information about
	// safely accessing the configuration.
	//
	// The configuration is represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config *DynamicValue

	// PriorPrivate is any provider-defined private state stored with the
	// resource. It is used for keeping state with the resource that is not
	// meant to be included when calculating diffs.
	PriorPrivate []byte

	// ProviderMeta supplies the provider metadata configuration for the
	// module this resource is in. Module-specific provider metadata is an
	// advanced feature and usage of it should be coordinated with the
	// Terraform Core team by raising an issue at
	// https://github.com/hashicorp/terraform/issues/new/choose. See the
	// documentation on `DynamicValue` for information about safely
	// accessing the configuration.
	//
	// The configuration is represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	//
	// This configuration will have known values for all fields.
	ProviderMeta *DynamicValue
}

// PlanResourceChangeResponse is the response from the provider about what the
// planned state for a given resource should be.
type PlanResourceChangeResponse struct {
	// PlannedState is the provider's indication of what the state for the
	// resource should be after apply, represented as a `DynamicValue`. See
	// the documentation for `DynamicValue` for information about safely
	// creating the `DynamicValue`.
	//
	// This is usually derived from the ProposedNewState passed in the
	// PlanResourceChangeRequest, with default values substituted for any
	// null values and overriding any computed values that are expected to
	// change as a result of the apply operation. This may contain unknown
	// values if the value could change but its new value won't be known
	// until apply time.
	//
	// Any value that was non-null in the configuration must either
	// preserve the exact configuration value or return the corresponding
	// value from the prior state. The value from the prior state should be
	// returned when the configuration value is semantically equivalent to
	// the state value.
	//
	// Any value that is marked as computed in the schema and is null in
	// the configuration may be set by the provider to any value of the
	// expected type.
	//
	// PlanResourceChange will actually be called twice; once when
	// generating the plan for the user to approve, once during the apply.
	// During the apply, additional values from the configuration--upstream
	// values interpolated in that were computed at apply time--will be
	// populated. During this second call, any attribute that had a known
	// value in the first PlannedState must have an identical value in the
	// second PlannedState. Any unknown values may remain unknown or may
	// take on any value of the appropriate type. This means the values
	// returned in PlannedState should be deterministic and unknown values
	// should be used if a field's value may change depending on what value
	// ends up filling an unknown value in the config.
	//
	// The state should be represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	PlannedState *DynamicValue

	// RequiresReplace is a list of tftypes.AttributePaths that require the
	// resource to be replaced. They should point to the specific field
	// that changed that requires the resource to be destroyed and
	// recreated.
	RequiresReplace []*tftypes.AttributePath

	// PlannedPrivate should be set to any state that the provider would
	// like sent with requests for this resource. This state will be
	// associated with the resource, but will not be considered when
	// calculating diffs.
	PlannedPrivate []byte

	// Diagnostics report errors or warnings related to determining the
	// planned state of the requested resource. Returning an empty slice
	// indicates a successful validation with no warnings or errors
	// generated.
	Diagnostics []*Diagnostic

	// UnsafeToUseLegacyTypeSystem should only be set by
	// hashicorp/terraform-plugin-sdk. It modifies Terraform's behavior to
	// work with the legacy expectations of that SDK.
	//
	// Nobody else should use this. Ever. For any reason. Just don't do it.
	//
	// We have to expose it here for terraform-plugin-sdk to be muxable, or
	// we wouldn't even be including it in this type. Don't use it. It may
	// go away or change behavior on you with no warning. It is
	// explicitly unsupported and not part of our SemVer guarantees.
	//
	// Deprecated: Really, just don't use this, you don't need it.
	UnsafeToUseLegacyTypeSystem bool
}

// ApplyResourceChangeRequest is the request Terraform sends when it needs to
// apply a planned set of changes to a resource.
type ApplyResourceChangeRequest struct {
	// TypeName is the type of resource Terraform wants to change.
	TypeName string

	// PriorState is the state of the resource before the changes are
	// applied, represented as a `DynamicValue`. See the documentation for
	// `DynamicValue` for information about safely accessing the state.
	//
	// The state is represented as a tftypes.Object, with each attribute
	// and nested block getting its own key and value.
	PriorState *DynamicValue

	// PlannedState is Terraform's plan for what the state should look like
	// after the changes are applied, represented as a `DynamicValue`. See
	// the documentation for `DynamicValue` for information about safely
	// accessing the state.
	//
	// This is the PlannedState returned during PlanResourceChange.
	//
	// The state is represented as a tftypes.Object, with each attribute
	// and nested block getting its own key and value.
	PlannedState *DynamicValue

	// Config is the configuration the user supplied for the resource. See
	// the documentation on `DynamicValue` for more information about
	// safely accessing the configuration.
	//
	// The configuration is represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	//
	// This configuration may contain unknown values.
	Config *DynamicValue

	// PlannedPrivate is any provider-defined private state stored with the
	// resource. It is used for keeping state with the resource that is not
	// meant to be included when calculating diffs.
	PlannedPrivate []byte

	// ProviderMeta supplies the provider metadata configuration for the
	// module this resource is in. Module-specific provider metadata is an
	// advanced feature and usage of it should be coordinated with the
	// Terraform Core team by raising an issue at
	// https://github.com/hashicorp/terraform/issues/new/choose. See the
	// documentation on `DynamicValue` for information about safely
	// accessing the configuration.
	//
	// The configuration is represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	//
	// This configuration will have known values for all fields.
	ProviderMeta *DynamicValue
}

// ApplyResourceChangeResponse is the response from the provider about what the
// state of a resource is after planned changes have been applied.
type ApplyResourceChangeResponse struct {
	// NewState is the provider's understanding of what the resource's
	// state is after changes are applied, represented as a `DynamicValue`.
	// See the documentation for `DynamicValue` for information about
	// safely creating the `DynamicValue`.
	//
	// Any attribute, whether computed or not, that has a known value in
	// the PlannedState in the ApplyResourceChangeRequest must be preserved
	// exactly as it was in NewState.
	//
	// Any attribute in the PlannedState in the ApplyResourceChangeRequest
	// that is unknown must take on a known value at this time. No unknown
	// values are allowed in the NewState.
	//
	// The state should be represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	NewState *DynamicValue

	// Private should be set to any state that the provider would like sent
	// with requests for this resource. This state will be associated with
	// the resource, but will not be considered when calculating diffs.
	Private []byte

	// Diagnostics report errors or warnings related to applying changes to
	// the requested resource. Returning an empty slice indicates a
	// successful validation with no warnings or errors generated.
	Diagnostics []*Diagnostic

	// UnsafeToUseLegacyTypeSystem should only be set by
	// hashicorp/terraform-plugin-sdk. It modifies Terraform's behavior to
	// work with the legacy expectations of that SDK.
	//
	// Nobody else should use this. Ever. For any reason. Just don't do it.
	//
	// We have to expose it here for terraform-plugin-sdk to be muxable, or
	// we wouldn't even be including it in this type. Don't use it. It may
	// go away or change behavior on you with no warning. It is
	// explicitly unsupported and not part of our SemVer guarantees.
	//
	// Deprecated: Really, just don't use this, you don't need it.
	UnsafeToUseLegacyTypeSystem bool
}

// ImportResourceStateRequest is the request Terraform sends when it wants a
// provider to import one or more resources specified by an ID.
type ImportResourceStateRequest struct {
	// TypeName is the type of resource Terraform wants to import.
	TypeName string

	// ID is the user-supplied identifying information about the resource
	// or resources. Providers decide and communicate to users the format
	// for the ID, and use it to determine what resource or resources to
	// import.
	ID string
}

// ImportResourceStateResponse is the response from the provider about the
// imported resources.
type ImportResourceStateResponse struct {
	// ImportedResources are the resources the provider found and was able
	// to import.
	ImportedResources []*ImportedResource

	// Diagnostics report errors or warnings related to importing the
	// requested resource or resources. Returning an empty slice indicates
	// a successful validation with no warnings or errors generated.
	Diagnostics []*Diagnostic
}

// ImportedResource represents a single resource that a provider has
// successfully imported into state.
type ImportedResource struct {
	// TypeName is the type of resource that was imported.
	TypeName string

	// State is the provider's understanding of the imported resource's
	// state, represented as a `DynamicValue`. See the documentation for
	// `DynamicValue` for information about safely creating the
	// `DynamicValue`.
	//
	// The state should be represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	State *DynamicValue

	// Private should be set to any state that the provider would like sent
	// with requests for this resource. This state will be associated with
	// the resource, but will not be considered when calculating diffs.
	Private []byte
}

package tfprotov6

import (
	"context"
)

// DataSourceServer is an interface containing the methods a data source
// implementation needs to fill.
type DataSourceServer interface {
	// ValidateDataResourceConfig is called when Terraform is checking that a
	// data source's configuration is valid. It is guaranteed to have types
	// conforming to your schema, but it is not guaranteed that all values
	// will be known. This is your opportunity to do custom or advanced
	// validation prior to a plan being generated.
	ValidateDataResourceConfig(context.Context, *ValidateDataResourceConfigRequest) (*ValidateDataResourceConfigResponse, error)

	// ReadDataSource is called when Terraform is refreshing a data
	// source's state.
	ReadDataSource(context.Context, *ReadDataSourceRequest) (*ReadDataSourceResponse, error)
}

// ValidateDataResourceConfigRequest is the request Terraform sends when it wants
// to validate a data source's configuration.
type ValidateDataResourceConfigRequest struct {
	// TypeName is the type of data source Terraform is validating.
	TypeName string

	// Config is the configuration the user supplied for that data source.
	// See the documentation on `DynamicValue` for more information about
	// safely accessing the configuration.
	//
	// The configuration is represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config *DynamicValue
}

// ValidateDataResourceConfigResponse is the response from the provider about the
// validity of a data source's configuration.
type ValidateDataResourceConfigResponse struct {
	// Diagnostics report errors or warnings related to the given
	// configuration. Returning an empty slice indicates a successful
	// validation with no warnings or errors generated.
	Diagnostics []*Diagnostic
}

// ReadDataSourceRequest is the request Terraform sends when it wants to get
// the latest state for a data source.
type ReadDataSourceRequest struct {
	// TypeName is the type of data source Terraform is requesting an
	// updated state for.
	TypeName string

	// Config is the configuration the user supplied for that data source.
	// See the documentation on `DynamicValue` for information about safely
	// accessing the configuration.
	//
	// The configuration is represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	//
	// This configuration may have unknown values.
	Config *DynamicValue

	// ProviderMeta supplies the provider metadata configuration for the
	// module this data source is in. Module-specific provider metadata is
	// an advanced feature and usage of it should be coordinated with the
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

// ReadDataSourceResponse is the response from the provider about the current
// state of the requested data source.
type ReadDataSourceResponse struct {
	// State is the current state of the data source, represented as a
	// `DynamicValue`. See the documentation on `DynamicValue` for
	// information about safely creating the `DynamicValue`.
	//
	// The state should be represented as a tftypes.Object, with each
	// attribute and nested block getting its own key and value.
	State *DynamicValue

	// Diagnostics report errors or warnings related to retrieving the
	// current state of the requested data source. Returning an empty slice
	// indicates a successful validation with no warnings or errors
	// generated.
	Diagnostics []*Diagnostic
}

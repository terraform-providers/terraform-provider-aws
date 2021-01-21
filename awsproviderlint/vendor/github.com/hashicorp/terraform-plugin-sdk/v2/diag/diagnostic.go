package diag

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
)

// Diagnostics is a collection of Diagnostic.
//
// Developers should append and build the list of diagnostics up until a fatal
// error is reached, at which point they should return the Diagnostics to the
// SDK.
type Diagnostics []Diagnostic

// HasError returns true is Diagnostics contains an instance of
// Severity == Error.
//
// This helper aims to mimic the go error practices of if err != nil. After any
// operation that returns Diagnostics, check that it HasError and bubble up the
// stack.
func (diags Diagnostics) HasError() bool {
	for i := range diags {
		if diags[i].Severity == Error {
			return true
		}
	}
	return false
}

// Diagnostic is a contextual message intended at outlining problems in user
// configuration.
//
// It supports multiple levels of severity (Error or Warning), a short Summary
// of the problem, an optional longer Detail message that can assist the user in
// fixing the problem, as well as an AttributePath representation which
// Terraform uses to indicate where the issue took place in the user's
// configuration.
//
// A Diagnostic will typically be used to pinpoint a problem with user
// configuration, however it can still be used to present warnings or errors
// to the user without any AttributePath set.
type Diagnostic struct {
	// Severity indicates the level of the Diagnostic. Currently can be set to
	// either Error or Warning
	Severity Severity

	// Summary is a short description of the problem, rendered above location
	// information
	Summary string

	// Detail is an optional second message rendered below location information
	// typically used to communicate a potential fix to the user.
	Detail string

	// AttributePath is a representation of the path starting from the root of
	// block (resource, datasource, provider) under evaluation by the SDK, to
	// the attribute that the Diagnostic should be associated to. Terraform will
	// use this information to render information on where the problem took
	// place in the user's configuration.
	//
	// It is represented with cty.Path, which is a list of steps of either
	// cty.GetAttrStep (an actual attribute) or cty.IndexStep (a step with Key
	// of cty.StringVal for map indexes, and cty.NumberVal for list indexes).
	//
	// PLEASE NOTE: While cty can support indexing into sets, the SDK and
	// protocol currently do not. For any Diagnostic related to a schema.TypeSet
	// or a child of that type, please terminate the path at the schema.TypeSet
	// and opt for more verbose Summary and Detail to help guide the user.
	//
	// Validity of the AttributePath is currently the responsibility of the
	// developer, Terraform should render the root block (provider, resource,
	// datasource) in cases where the attribute path is invalid.
	AttributePath cty.Path
}

// Validate ensures a valid Severity and a non-empty Summary are set.
func (d Diagnostic) Validate() error {
	var validSev bool
	for _, sev := range severities {
		if d.Severity == sev {
			validSev = true
			break
		}
	}
	if !validSev {
		return fmt.Errorf("invalid severity: %v", d.Severity)
	}
	if d.Summary == "" {
		return errors.New("empty summary")
	}
	return nil
}

// Severity is an enum type marking the severity level of a Diagnostic
type Severity int

const (
	Error Severity = iota
	Warning
)

var severities = []Severity{Error, Warning}

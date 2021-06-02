package tfprotov5

import "github.com/hashicorp/terraform-plugin-go/tftypes"

const (
	// DiagnosticSeverityInvalid is used to indicate an invalid
	// `DiagnosticSeverity`. Provider developers should not use it.
	DiagnosticSeverityInvalid DiagnosticSeverity = 0

	// DiagnosticSeverityError is used to indicate that a `Diagnostic`
	// represents an error and should halt Terraform execution.
	DiagnosticSeverityError DiagnosticSeverity = 1

	// DiagnosticSeverityWarning is used to indicate that a `Diagnostic`
	// represents a warning and should not halt Terraform's execution, but
	// it should be surfaced to the user.
	DiagnosticSeverityWarning DiagnosticSeverity = 2
)

// Diagnostic is used to convey information back the user running Terraform.
type Diagnostic struct {
	// Severity indicates how Terraform should handle the Diagnostic.
	Severity DiagnosticSeverity

	// Summary is a brief description of the problem, roughly
	// sentence-sized, and should provide a concise description of what
	// went wrong. For example, a Summary could be as simple as "Invalid
	// value.".
	Summary string

	// Detail is a lengthier, more complete description of the problem.
	// Detail should provide enough information that a user can resolve the
	// problem entirely. For example, a Detail could be "Values must be
	// alphanumeric and lowercase only."
	Detail string

	// Attribute indicates which field, specifically, has the problem. Not
	// setting this will indicate the entire resource; setting it will
	// indicate that the problem is with a certain field in the resource,
	// which helps users find the source of the problem.
	Attribute *tftypes.AttributePath
}

// DiagnosticSeverity represents different classes of Diagnostic which affect
// how Terraform handles the Diagnostics.
type DiagnosticSeverity int32

func (d DiagnosticSeverity) String() string {
	switch d {
	case 0:
		return "INVALID"
	case 1:
		return "ERROR"
	case 2:
		return "WARNING"
	}
	return "UNKNOWN"
}

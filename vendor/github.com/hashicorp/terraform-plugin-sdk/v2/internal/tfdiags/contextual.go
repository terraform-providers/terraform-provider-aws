package tfdiags

import (
	"github.com/hashicorp/go-cty/cty"
)

// AttributeValue returns a diagnostic about an attribute value in an implied current
// configuration context. This should be returned only from functions whose
// interface specifies a clear configuration context that this will be
// resolved in.
//
// The given path is relative to the implied configuration context. To describe
// a top-level attribute, it should be a single-element cty.Path with a
// cty.GetAttrStep. It's assumed that the path is returning into a structure
// that would be produced by our conventions in the configschema package; it
// may return unexpected results for structures that can't be represented by
// configschema.
//
// Since mapping attribute paths back onto configuration is an imprecise
// operation (e.g. dynamic block generation may cause the same block to be
// evaluated multiple times) the diagnostic detail should include the attribute
// name and other context required to help the user understand what is being
// referenced in case the identified source range is not unique.
//
// The returned attribute will not have source location information until
// context is applied to the containing diagnostics using diags.InConfigBody.
// After context is applied, the source location is the value assigned to the
// named attribute, or the containing body's "missing item range" if no
// value is present.
func AttributeValue(severity Severity, summary, detail string, attrPath cty.Path) Diagnostic {
	return &attributeDiagnostic{
		diagnosticBase: diagnosticBase{
			severity: severity,
			summary:  summary,
			detail:   detail,
		},
		attrPath: attrPath,
	}
}

// GetAttribute extracts an attribute cty.Path from a diagnostic if it contains
// one. Normally this is not accessed directly, and instead the config body is
// added to the Diagnostic to create a more complete message for the user. In
// some cases however, we may want to know just the name of the attribute that
// generated the Diagnostic message.
// This returns a nil cty.Path if it does not exist in the Diagnostic.
func GetAttribute(d Diagnostic) cty.Path {
	if d, ok := d.(*attributeDiagnostic); ok {
		return d.attrPath
	}
	return nil
}

type attributeDiagnostic struct {
	diagnosticBase
	attrPath cty.Path
}

// WholeContainingBody returns a diagnostic about the body that is an implied
// current configuration context. This should be returned only from
// functions whose interface specifies a clear configuration context that this
// will be resolved in.
//
// The returned attribute will not have source location information until
// context is applied to the containing diagnostics using diags.InConfigBody.
// After context is applied, the source location is currently the missing item
// range of the body. In future, this may change to some other suitable
// part of the containing body.
func WholeContainingBody(severity Severity, summary, detail string) Diagnostic {
	return &wholeBodyDiagnostic{
		diagnosticBase: diagnosticBase{
			severity: severity,
			summary:  summary,
			detail:   detail,
		},
	}
}

type wholeBodyDiagnostic struct {
	diagnosticBase
}

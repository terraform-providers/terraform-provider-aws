package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func Diagnostic(in *tfplugin6.Diagnostic) (*tfprotov6.Diagnostic, error) {
	diag := &tfprotov6.Diagnostic{
		Severity: DiagnosticSeverity(in.Severity),
		Summary:  in.Summary,
		Detail:   in.Detail,
	}
	if in.Attribute != nil {
		attr, err := AttributePath(in.Attribute)
		if err != nil {
			return diag, err
		}
		diag.Attribute = attr
	}
	return diag, nil
}

func DiagnosticSeverity(in tfplugin6.Diagnostic_Severity) tfprotov6.DiagnosticSeverity {
	return tfprotov6.DiagnosticSeverity(in)
}

func Diagnostics(in []*tfplugin6.Diagnostic) ([]*tfprotov6.Diagnostic, error) {
	diagnostics := make([]*tfprotov6.Diagnostic, 0, len(in))
	for _, diag := range in {
		if diag == nil {
			diagnostics = append(diagnostics, nil)
			continue
		}
		d, err := Diagnostic(diag)
		if err != nil {
			return diagnostics, err
		}
		diagnostics = append(diagnostics, d)
	}
	return diagnostics, nil
}

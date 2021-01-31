package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func Diagnostic(in *tfplugin5.Diagnostic) (*tfprotov5.Diagnostic, error) {
	diag := &tfprotov5.Diagnostic{
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

func DiagnosticSeverity(in tfplugin5.Diagnostic_Severity) tfprotov5.DiagnosticSeverity {
	return tfprotov5.DiagnosticSeverity(in)
}

func Diagnostics(in []*tfplugin5.Diagnostic) ([]*tfprotov5.Diagnostic, error) {
	diagnostics := make([]*tfprotov5.Diagnostic, 0, len(in))
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

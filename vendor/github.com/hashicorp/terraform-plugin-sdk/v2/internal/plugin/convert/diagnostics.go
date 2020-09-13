package convert

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	proto "github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfplugin5"
)

// AppendProtoDiag appends a new diagnostic from a warning string or an error.
// This panics if d is not a string or error.
func AppendProtoDiag(diags []*proto.Diagnostic, d interface{}) []*proto.Diagnostic {
	switch d := d.(type) {
	case cty.PathError:
		ap := PathToAttributePath(d.Path)
		diags = append(diags, &proto.Diagnostic{
			Severity:  proto.Diagnostic_ERROR,
			Summary:   d.Error(),
			Attribute: ap,
		})
	case diag.Diagnostics:
		diags = append(diags, DiagsToProto(d)...)
	case error:
		diags = append(diags, &proto.Diagnostic{
			Severity: proto.Diagnostic_ERROR,
			Summary:  d.Error(),
		})
	case string:
		diags = append(diags, &proto.Diagnostic{
			Severity: proto.Diagnostic_WARNING,
			Summary:  d,
		})
	case *proto.Diagnostic:
		diags = append(diags, d)
	case []*proto.Diagnostic:
		diags = append(diags, d...)
	}
	return diags
}

// ProtoToDiags converts a list of proto.Diagnostics to a diag.Diagnostics.
func ProtoToDiags(ds []*proto.Diagnostic) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, d := range ds {
		var severity diag.Severity

		switch d.Severity {
		case proto.Diagnostic_ERROR:
			severity = diag.Error
		case proto.Diagnostic_WARNING:
			severity = diag.Warning
		}

		diags = append(diags, diag.Diagnostic{
			Severity:      severity,
			Summary:       d.Summary,
			Detail:        d.Detail,
			AttributePath: AttributePathToPath(d.Attribute),
		})
	}

	return diags
}

func DiagsToProto(diags diag.Diagnostics) []*proto.Diagnostic {
	var ds []*proto.Diagnostic
	for _, d := range diags {
		if err := d.Validate(); err != nil {
			panic(fmt.Errorf("Invalid diagnostic: %s. This is always a bug in the provider implementation", err))
		}
		protoDiag := &proto.Diagnostic{
			Summary:   d.Summary,
			Detail:    d.Detail,
			Attribute: PathToAttributePath(d.AttributePath),
		}
		if d.Severity == diag.Error {
			protoDiag.Severity = proto.Diagnostic_ERROR
		} else if d.Severity == diag.Warning {
			protoDiag.Severity = proto.Diagnostic_WARNING
		}
		ds = append(ds, protoDiag)
	}
	return ds
}

// AttributePathToPath takes the proto encoded path and converts it to a cty.Path
func AttributePathToPath(ap *proto.AttributePath) cty.Path {
	var p cty.Path
	if ap == nil {
		return p
	}
	for _, step := range ap.Steps {
		switch selector := step.Selector.(type) {
		case *proto.AttributePath_Step_AttributeName:
			p = p.GetAttr(selector.AttributeName)
		case *proto.AttributePath_Step_ElementKeyString:
			p = p.Index(cty.StringVal(selector.ElementKeyString))
		case *proto.AttributePath_Step_ElementKeyInt:
			p = p.Index(cty.NumberIntVal(selector.ElementKeyInt))
		}
	}
	return p
}

// PathToAttributePath takes a cty.Path and converts it to a proto-encoded path.
func PathToAttributePath(p cty.Path) *proto.AttributePath {
	ap := &proto.AttributePath{}
	for _, step := range p {
		switch selector := step.(type) {
		case cty.GetAttrStep:
			ap.Steps = append(ap.Steps, &proto.AttributePath_Step{
				Selector: &proto.AttributePath_Step_AttributeName{
					AttributeName: selector.Name,
				},
			})
		case cty.IndexStep:
			key := selector.Key
			switch key.Type() {
			case cty.String:
				ap.Steps = append(ap.Steps, &proto.AttributePath_Step{
					Selector: &proto.AttributePath_Step_ElementKeyString{
						ElementKeyString: key.AsString(),
					},
				})
			case cty.Number:
				v, _ := key.AsBigFloat().Int64()
				ap.Steps = append(ap.Steps, &proto.AttributePath_Step{
					Selector: &proto.AttributePath_Step_ElementKeyInt{
						ElementKeyInt: v,
					},
				})
			default:
				// We'll bail early if we encounter anything else, and just
				// return the valid prefix.
				return ap
			}
		}
	}
	return ap
}

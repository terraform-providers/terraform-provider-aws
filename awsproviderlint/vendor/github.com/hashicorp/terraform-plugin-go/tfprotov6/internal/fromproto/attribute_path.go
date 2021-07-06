package fromproto

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var ErrUnknownAttributePathStepType = errors.New("unknown type of AttributePath_Step")

func AttributePath(in *tfplugin6.AttributePath) (*tftypes.AttributePath, error) {
	steps, err := AttributePathSteps(in.Steps)
	if err != nil {
		return nil, err
	}
	return tftypes.NewAttributePathWithSteps(steps), nil
}

func AttributePaths(in []*tfplugin6.AttributePath) ([]*tftypes.AttributePath, error) {
	resp := make([]*tftypes.AttributePath, 0, len(in))
	for _, a := range in {
		if a == nil {
			resp = append(resp, nil)
			continue
		}
		attr, err := AttributePath(a)
		if err != nil {
			return resp, err
		}
		resp = append(resp, attr)
	}
	return resp, nil
}

func AttributePathStep(step *tfplugin6.AttributePath_Step) (tftypes.AttributePathStep, error) {
	selector := step.GetSelector()
	if v, ok := selector.(*tfplugin6.AttributePath_Step_AttributeName); ok {
		return tftypes.AttributeName(v.AttributeName), nil
	}
	if v, ok := selector.(*tfplugin6.AttributePath_Step_ElementKeyString); ok {
		return tftypes.ElementKeyString(v.ElementKeyString), nil
	}
	if v, ok := selector.(*tfplugin6.AttributePath_Step_ElementKeyInt); ok {
		return tftypes.ElementKeyInt(v.ElementKeyInt), nil
	}
	return nil, ErrUnknownAttributePathStepType
}

func AttributePathSteps(in []*tfplugin6.AttributePath_Step) ([]tftypes.AttributePathStep, error) {
	resp := make([]tftypes.AttributePathStep, 0, len(in))
	for _, step := range in {
		if step == nil {
			continue
		}
		s, err := AttributePathStep(step)
		if err != nil {
			return resp, err
		}
		resp = append(resp, s)
	}
	return resp, nil
}

package fromproto

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func ValidateResourceConfigRequest(in *tfplugin6.ValidateResourceConfig_Request) (*tfprotov6.ValidateResourceConfigRequest, error) {
	resp := &tfprotov6.ValidateResourceConfigRequest{
		TypeName: in.TypeName,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return resp, nil
}

func ValidateResourceConfigResponse(in *tfplugin6.ValidateResourceConfig_Response) (*tfprotov6.ValidateResourceConfigResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfprotov6.ValidateResourceConfigResponse{
		Diagnostics: diags,
	}, nil
}

func UpgradeResourceStateRequest(in *tfplugin6.UpgradeResourceState_Request) (*tfprotov6.UpgradeResourceStateRequest, error) {
	resp := &tfprotov6.UpgradeResourceStateRequest{
		TypeName: in.TypeName,
		Version:  in.Version,
	}
	if in.RawState != nil {
		resp.RawState = RawState(in.RawState)
	}
	return resp, nil
}

func UpgradeResourceStateResponse(in *tfplugin6.UpgradeResourceState_Response) (*tfprotov6.UpgradeResourceStateResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	resp := &tfprotov6.UpgradeResourceStateResponse{
		Diagnostics: diags,
	}
	if in.UpgradedState != nil {
		resp.UpgradedState = DynamicValue(in.UpgradedState)
	}
	return resp, nil
}

func ReadResourceRequest(in *tfplugin6.ReadResource_Request) (*tfprotov6.ReadResourceRequest, error) {
	resp := &tfprotov6.ReadResourceRequest{
		TypeName: in.TypeName,
		Private:  in.Private,
	}
	if in.CurrentState != nil {
		resp.CurrentState = DynamicValue(in.CurrentState)
	}
	if in.ProviderMeta != nil {
		resp.ProviderMeta = DynamicValue(in.ProviderMeta)
	}
	return resp, nil
}

func ReadResourceResponse(in *tfplugin6.ReadResource_Response) (*tfprotov6.ReadResourceResponse, error) {
	resp := &tfprotov6.ReadResourceResponse{
		Private: in.Private,
	}
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return resp, err
	}
	resp.Diagnostics = diags
	if in.NewState != nil {
		resp.NewState = DynamicValue(in.NewState)
	}
	return resp, nil
}

func PlanResourceChangeRequest(in *tfplugin6.PlanResourceChange_Request) (*tfprotov6.PlanResourceChangeRequest, error) {
	resp := &tfprotov6.PlanResourceChangeRequest{
		TypeName:     in.TypeName,
		PriorPrivate: in.PriorPrivate,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	if in.PriorState != nil {
		resp.PriorState = DynamicValue(in.PriorState)
	}
	if in.ProposedNewState != nil {
		resp.ProposedNewState = DynamicValue(in.ProposedNewState)
	}
	if in.ProviderMeta != nil {
		resp.ProviderMeta = DynamicValue(in.ProviderMeta)
	}
	return resp, nil
}

func PlanResourceChangeResponse(in *tfplugin6.PlanResourceChange_Response) (*tfprotov6.PlanResourceChangeResponse, error) {
	resp := &tfprotov6.PlanResourceChangeResponse{
		PlannedPrivate: in.PlannedPrivate,
	}
	attributePaths, err := AttributePaths(in.RequiresReplace)
	if err != nil {
		return resp, err
	}
	resp.RequiresReplace = attributePaths
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return resp, err
	}
	resp.Diagnostics = diags
	if in.PlannedState != nil {
		resp.PlannedState = DynamicValue(in.PlannedState)
	}
	return resp, nil
}

func ApplyResourceChangeRequest(in *tfplugin6.ApplyResourceChange_Request) (*tfprotov6.ApplyResourceChangeRequest, error) {
	resp := &tfprotov6.ApplyResourceChangeRequest{
		TypeName:       in.TypeName,
		PlannedPrivate: in.PlannedPrivate,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	if in.PriorState != nil {
		resp.PriorState = DynamicValue(in.PriorState)
	}
	if in.PlannedState != nil {
		resp.PlannedState = DynamicValue(in.PlannedState)
	}
	if in.ProviderMeta != nil {
		resp.ProviderMeta = DynamicValue(in.ProviderMeta)
	}
	return resp, nil
}

func ApplyResourceChangeResponse(in *tfplugin6.ApplyResourceChange_Response) (*tfprotov6.ApplyResourceChangeResponse, error) {
	resp := &tfprotov6.ApplyResourceChangeResponse{
		Private: in.Private,
	}
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return resp, err
	}
	resp.Diagnostics = diags
	if in.NewState != nil {
		resp.NewState = DynamicValue(in.NewState)
	}
	return resp, nil
}

func ImportResourceStateRequest(in *tfplugin6.ImportResourceState_Request) (*tfprotov6.ImportResourceStateRequest, error) {
	return &tfprotov6.ImportResourceStateRequest{
		TypeName: in.TypeName,
		ID:       in.Id,
	}, nil
}

func ImportResourceStateResponse(in *tfplugin6.ImportResourceState_Response) (*tfprotov6.ImportResourceStateResponse, error) {
	imported, err := ImportedResources(in.ImportedResources)
	if err != nil {
		return nil, err
	}
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfprotov6.ImportResourceStateResponse{
		ImportedResources: imported,
		Diagnostics:       diags,
	}, nil
}

func ImportedResource(in *tfplugin6.ImportResourceState_ImportedResource) (*tfprotov6.ImportedResource, error) {
	resp := &tfprotov6.ImportedResource{
		TypeName: in.TypeName,
		Private:  in.Private,
	}
	if in.State != nil {
		resp.State = DynamicValue(in.State)
	}
	return resp, nil
}

func ImportedResources(in []*tfplugin6.ImportResourceState_ImportedResource) ([]*tfprotov6.ImportedResource, error) {
	resp := make([]*tfprotov6.ImportedResource, 0, len(in))
	for pos, i := range in {
		if i == nil {
			resp = append(resp, nil)
			continue
		}
		r, err := ImportedResource(i)
		if err != nil {
			return resp, fmt.Errorf("Error converting imported resource %d/%d: %w", pos+1, len(in), err)
		}
		resp = append(resp, r)
	}
	return resp, nil
}

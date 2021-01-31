package fromproto

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func ValidateResourceTypeConfigRequest(in *tfplugin5.ValidateResourceTypeConfig_Request) (*tfprotov5.ValidateResourceTypeConfigRequest, error) {
	resp := &tfprotov5.ValidateResourceTypeConfigRequest{
		TypeName: in.TypeName,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return resp, nil
}

func ValidateResourceTypeConfigResponse(in *tfplugin5.ValidateResourceTypeConfig_Response) (*tfprotov5.ValidateResourceTypeConfigResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfprotov5.ValidateResourceTypeConfigResponse{
		Diagnostics: diags,
	}, nil
}

func UpgradeResourceStateRequest(in *tfplugin5.UpgradeResourceState_Request) (*tfprotov5.UpgradeResourceStateRequest, error) {
	resp := &tfprotov5.UpgradeResourceStateRequest{
		TypeName: in.TypeName,
		Version:  in.Version,
	}
	if in.RawState != nil {
		resp.RawState = RawState(in.RawState)
	}
	return resp, nil
}

func UpgradeResourceStateResponse(in *tfplugin5.UpgradeResourceState_Response) (*tfprotov5.UpgradeResourceStateResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	resp := &tfprotov5.UpgradeResourceStateResponse{
		Diagnostics: diags,
	}
	if in.UpgradedState != nil {
		resp.UpgradedState = DynamicValue(in.UpgradedState)
	}
	return resp, nil
}

func ReadResourceRequest(in *tfplugin5.ReadResource_Request) (*tfprotov5.ReadResourceRequest, error) {
	resp := &tfprotov5.ReadResourceRequest{
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

func ReadResourceResponse(in *tfplugin5.ReadResource_Response) (*tfprotov5.ReadResourceResponse, error) {
	resp := &tfprotov5.ReadResourceResponse{
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

func PlanResourceChangeRequest(in *tfplugin5.PlanResourceChange_Request) (*tfprotov5.PlanResourceChangeRequest, error) {
	resp := &tfprotov5.PlanResourceChangeRequest{
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

func PlanResourceChangeResponse(in *tfplugin5.PlanResourceChange_Response) (*tfprotov5.PlanResourceChangeResponse, error) {
	resp := &tfprotov5.PlanResourceChangeResponse{
		PlannedPrivate:              in.PlannedPrivate,
		UnsafeToUseLegacyTypeSystem: in.LegacyTypeSystem,
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

func ApplyResourceChangeRequest(in *tfplugin5.ApplyResourceChange_Request) (*tfprotov5.ApplyResourceChangeRequest, error) {
	resp := &tfprotov5.ApplyResourceChangeRequest{
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

func ApplyResourceChangeResponse(in *tfplugin5.ApplyResourceChange_Response) (*tfprotov5.ApplyResourceChangeResponse, error) {
	resp := &tfprotov5.ApplyResourceChangeResponse{
		Private:                     in.Private,
		UnsafeToUseLegacyTypeSystem: in.LegacyTypeSystem,
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

func ImportResourceStateRequest(in *tfplugin5.ImportResourceState_Request) (*tfprotov5.ImportResourceStateRequest, error) {
	return &tfprotov5.ImportResourceStateRequest{
		TypeName: in.TypeName,
		ID:       in.Id,
	}, nil
}

func ImportResourceStateResponse(in *tfplugin5.ImportResourceState_Response) (*tfprotov5.ImportResourceStateResponse, error) {
	imported, err := ImportedResources(in.ImportedResources)
	if err != nil {
		return nil, err
	}
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfprotov5.ImportResourceStateResponse{
		ImportedResources: imported,
		Diagnostics:       diags,
	}, nil
}

func ImportedResource(in *tfplugin5.ImportResourceState_ImportedResource) (*tfprotov5.ImportedResource, error) {
	resp := &tfprotov5.ImportedResource{
		TypeName: in.TypeName,
		Private:  in.Private,
	}
	if in.State != nil {
		resp.State = DynamicValue(in.State)
	}
	return resp, nil
}

func ImportedResources(in []*tfplugin5.ImportResourceState_ImportedResource) ([]*tfprotov5.ImportedResource, error) {
	resp := make([]*tfprotov5.ImportedResource, 0, len(in))
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

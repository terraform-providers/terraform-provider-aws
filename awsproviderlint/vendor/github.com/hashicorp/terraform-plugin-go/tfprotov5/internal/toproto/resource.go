package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func ValidateResourceTypeConfig_Request(in *tfprotov5.ValidateResourceTypeConfigRequest) (*tfplugin5.ValidateResourceTypeConfig_Request, error) {
	resp := &tfplugin5.ValidateResourceTypeConfig_Request{
		TypeName: in.TypeName,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return resp, nil
}

func ValidateResourceTypeConfig_Response(in *tfprotov5.ValidateResourceTypeConfigResponse) (*tfplugin5.ValidateResourceTypeConfig_Response, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfplugin5.ValidateResourceTypeConfig_Response{
		Diagnostics: diags,
	}, nil
}

func UpgradeResourceState_Request(in *tfprotov5.UpgradeResourceStateRequest) (*tfplugin5.UpgradeResourceState_Request, error) {
	resp := &tfplugin5.UpgradeResourceState_Request{
		TypeName: in.TypeName,
		Version:  in.Version,
	}
	if in.RawState != nil {
		resp.RawState = RawState(in.RawState)
	}
	return resp, nil
}

func UpgradeResourceState_Response(in *tfprotov5.UpgradeResourceStateResponse) (*tfplugin5.UpgradeResourceState_Response, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	resp := &tfplugin5.UpgradeResourceState_Response{
		Diagnostics: diags,
	}
	if in.UpgradedState != nil {
		resp.UpgradedState = DynamicValue(in.UpgradedState)
	}
	return resp, nil
}

func ReadResource_Request(in *tfprotov5.ReadResourceRequest) (*tfplugin5.ReadResource_Request, error) {
	resp := &tfplugin5.ReadResource_Request{
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

func ReadResource_Response(in *tfprotov5.ReadResourceResponse) (*tfplugin5.ReadResource_Response, error) {
	resp := &tfplugin5.ReadResource_Response{
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

func PlanResourceChange_Request(in *tfprotov5.PlanResourceChangeRequest) (*tfplugin5.PlanResourceChange_Request, error) {
	resp := &tfplugin5.PlanResourceChange_Request{
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

func PlanResourceChange_Response(in *tfprotov5.PlanResourceChangeResponse) (*tfplugin5.PlanResourceChange_Response, error) {
	resp := &tfplugin5.PlanResourceChange_Response{
		PlannedPrivate:   in.PlannedPrivate,
		LegacyTypeSystem: in.UnsafeToUseLegacyTypeSystem, //nolint:staticcheck
	}
	requiresReplace, err := AttributePaths(in.RequiresReplace)
	if err != nil {
		return resp, err
	}
	resp.RequiresReplace = requiresReplace
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

func ApplyResourceChange_Request(in *tfprotov5.ApplyResourceChangeRequest) (*tfplugin5.ApplyResourceChange_Request, error) {
	resp := &tfplugin5.ApplyResourceChange_Request{
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

func ApplyResourceChange_Response(in *tfprotov5.ApplyResourceChangeResponse) (*tfplugin5.ApplyResourceChange_Response, error) {
	resp := &tfplugin5.ApplyResourceChange_Response{
		Private:          in.Private,
		LegacyTypeSystem: in.UnsafeToUseLegacyTypeSystem, //nolint:staticcheck
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

func ImportResourceState_Request(in *tfprotov5.ImportResourceStateRequest) (*tfplugin5.ImportResourceState_Request, error) {
	return &tfplugin5.ImportResourceState_Request{
		TypeName: in.TypeName,
		Id:       in.ID,
	}, nil
}

func ImportResourceState_Response(in *tfprotov5.ImportResourceStateResponse) (*tfplugin5.ImportResourceState_Response, error) {
	importedResources, err := ImportResourceState_ImportedResources(in.ImportedResources)
	if err != nil {
		return nil, err
	}
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfplugin5.ImportResourceState_Response{
		ImportedResources: importedResources,
		Diagnostics:       diags,
	}, nil
}

func ImportResourceState_ImportedResource(in *tfprotov5.ImportedResource) (*tfplugin5.ImportResourceState_ImportedResource, error) {
	resp := &tfplugin5.ImportResourceState_ImportedResource{
		TypeName: in.TypeName,
		Private:  in.Private,
	}
	if in.State != nil {
		resp.State = DynamicValue(in.State)
	}
	return resp, nil
}

func ImportResourceState_ImportedResources(in []*tfprotov5.ImportedResource) ([]*tfplugin5.ImportResourceState_ImportedResource, error) {
	resp := make([]*tfplugin5.ImportResourceState_ImportedResource, 0, len(in))
	for _, i := range in {
		if i == nil {
			resp = append(resp, nil)
			continue
		}
		r, err := ImportResourceState_ImportedResource(i)
		if err != nil {
			return resp, err
		}
		resp = append(resp, r)
	}
	return resp, nil
}

// we have to say this next thing to get golint to stop yelling at us about the
// underscores in the function names. We want the function names to match
// actually-generated code, so it feels like fair play. It's just a shame we
// lose golint for the entire file.
//
// This file is not actually generated. You can edit it. Ignore this next line.
// Code generated by hand ignore this next bit DO NOT EDIT.

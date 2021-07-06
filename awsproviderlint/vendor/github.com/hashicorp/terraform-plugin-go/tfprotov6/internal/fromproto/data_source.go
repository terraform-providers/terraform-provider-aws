package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func ValidateDataResourceConfigRequest(in *tfplugin6.ValidateDataResourceConfig_Request) (*tfprotov6.ValidateDataResourceConfigRequest, error) {
	resp := &tfprotov6.ValidateDataResourceConfigRequest{
		TypeName: in.TypeName,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return resp, nil
}

func ValidateDataResourceConfigResponse(in *tfplugin6.ValidateDataResourceConfig_Response) (*tfprotov6.ValidateDataResourceConfigResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfprotov6.ValidateDataResourceConfigResponse{
		Diagnostics: diags,
	}, nil
}

func ReadDataSourceRequest(in *tfplugin6.ReadDataSource_Request) (*tfprotov6.ReadDataSourceRequest, error) {
	resp := &tfprotov6.ReadDataSourceRequest{
		TypeName: in.TypeName,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	if in.ProviderMeta != nil {
		resp.ProviderMeta = DynamicValue(in.ProviderMeta)
	}
	return resp, nil
}

func ReadDataSourceResponse(in *tfplugin6.ReadDataSource_Response) (*tfprotov6.ReadDataSourceResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	resp := &tfprotov6.ReadDataSourceResponse{
		Diagnostics: diags,
	}
	if in.State != nil {
		resp.State = DynamicValue(in.State)
	}
	return resp, nil
}

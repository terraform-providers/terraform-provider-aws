package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func ValidateDataSourceConfigRequest(in *tfplugin5.ValidateDataSourceConfig_Request) (*tfprotov5.ValidateDataSourceConfigRequest, error) {
	resp := &tfprotov5.ValidateDataSourceConfigRequest{
		TypeName: in.TypeName,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return resp, nil
}

func ValidateDataSourceConfigResponse(in *tfplugin5.ValidateDataSourceConfig_Response) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfprotov5.ValidateDataSourceConfigResponse{
		Diagnostics: diags,
	}, nil
}

func ReadDataSourceRequest(in *tfplugin5.ReadDataSource_Request) (*tfprotov5.ReadDataSourceRequest, error) {
	resp := &tfprotov5.ReadDataSourceRequest{
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

func ReadDataSourceResponse(in *tfplugin5.ReadDataSource_Response) (*tfprotov5.ReadDataSourceResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	resp := &tfprotov5.ReadDataSourceResponse{
		Diagnostics: diags,
	}
	if in.State != nil {
		resp.State = DynamicValue(in.State)
	}
	return resp, nil
}

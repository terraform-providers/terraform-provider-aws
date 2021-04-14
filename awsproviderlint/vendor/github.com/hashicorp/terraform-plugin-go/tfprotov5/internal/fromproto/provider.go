package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func GetProviderSchemaRequest(in *tfplugin5.GetProviderSchema_Request) (*tfprotov5.GetProviderSchemaRequest, error) {
	return &tfprotov5.GetProviderSchemaRequest{}, nil
}

func GetProviderSchemaResponse(in *tfplugin5.GetProviderSchema_Response) (*tfprotov5.GetProviderSchemaResponse, error) {
	var resp tfprotov5.GetProviderSchemaResponse
	if in.Provider != nil {
		schema, err := Schema(in.Provider)
		if err != nil {
			return &resp, err
		}
		resp.Provider = schema
	}
	if in.ProviderMeta != nil {
		schema, err := Schema(in.ProviderMeta)
		if err != nil {
			return &resp, err
		}
		resp.ProviderMeta = schema
	}
	resp.ResourceSchemas = make(map[string]*tfprotov5.Schema, len(in.ResourceSchemas))
	for k, v := range in.ResourceSchemas {
		if v == nil {
			resp.ResourceSchemas[k] = nil
			continue
		}
		schema, err := Schema(v)
		if err != nil {
			return &resp, err
		}
		resp.ResourceSchemas[k] = schema
	}
	resp.DataSourceSchemas = make(map[string]*tfprotov5.Schema, len(in.DataSourceSchemas))
	for k, v := range in.DataSourceSchemas {
		if v == nil {
			resp.DataSourceSchemas[k] = nil
			continue
		}
		schema, err := Schema(v)
		if err != nil {
			return &resp, err
		}
		resp.DataSourceSchemas[k] = schema
	}
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return &resp, err
	}
	resp.Diagnostics = diags
	return &resp, nil
}

func PrepareProviderConfigRequest(in *tfplugin5.PrepareProviderConfig_Request) (*tfprotov5.PrepareProviderConfigRequest, error) {
	var resp tfprotov5.PrepareProviderConfigRequest
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return &resp, nil
}

func PrepareProviderConfigResponse(in *tfplugin5.PrepareProviderConfig_Response) (*tfprotov5.PrepareProviderConfigResponse, error) {
	var resp tfprotov5.PrepareProviderConfigResponse
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	resp.Diagnostics = diags
	if in.PreparedConfig != nil {
		resp.PreparedConfig = DynamicValue(in.PreparedConfig)
	}
	return &resp, nil
}

func ConfigureProviderRequest(in *tfplugin5.Configure_Request) (*tfprotov5.ConfigureProviderRequest, error) {
	resp := &tfprotov5.ConfigureProviderRequest{
		TerraformVersion: in.TerraformVersion,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return resp, nil
}

func ConfigureProviderResponse(in *tfplugin5.Configure_Response) (*tfprotov5.ConfigureProviderResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfprotov5.ConfigureProviderResponse{
		Diagnostics: diags,
	}, nil
}

func StopProviderRequest(in *tfplugin5.Stop_Request) (*tfprotov5.StopProviderRequest, error) {
	return &tfprotov5.StopProviderRequest{}, nil
}

func StopProviderResponse(in *tfplugin5.Stop_Response) (*tfprotov5.StopProviderResponse, error) {
	return &tfprotov5.StopProviderResponse{
		Error: in.Error,
	}, nil
}

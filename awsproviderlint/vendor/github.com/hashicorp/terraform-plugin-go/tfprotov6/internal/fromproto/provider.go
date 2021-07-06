package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func GetProviderSchemaRequest(in *tfplugin6.GetProviderSchema_Request) (*tfprotov6.GetProviderSchemaRequest, error) {
	return &tfprotov6.GetProviderSchemaRequest{}, nil
}

func GetProviderSchemaResponse(in *tfplugin6.GetProviderSchema_Response) (*tfprotov6.GetProviderSchemaResponse, error) {
	var resp tfprotov6.GetProviderSchemaResponse
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
	resp.ResourceSchemas = make(map[string]*tfprotov6.Schema, len(in.ResourceSchemas))
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
	resp.DataSourceSchemas = make(map[string]*tfprotov6.Schema, len(in.DataSourceSchemas))
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

func ValidateProviderConfigRequest(in *tfplugin6.ValidateProviderConfig_Request) (*tfprotov6.ValidateProviderConfigRequest, error) {
	var resp tfprotov6.ValidateProviderConfigRequest
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return &resp, nil
}

func ValidateProviderConfigResponse(in *tfplugin6.ValidateProviderConfig_Response) (*tfprotov6.ValidateProviderConfigResponse, error) {
	var resp tfprotov6.ValidateProviderConfigResponse
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	resp.Diagnostics = diags
	return &resp, nil
}

func ConfigureProviderRequest(in *tfplugin6.ConfigureProvider_Request) (*tfprotov6.ConfigureProviderRequest, error) {
	resp := &tfprotov6.ConfigureProviderRequest{
		TerraformVersion: in.TerraformVersion,
	}
	if in.Config != nil {
		resp.Config = DynamicValue(in.Config)
	}
	return resp, nil
}

func ConfigureProviderResponse(in *tfplugin6.ConfigureProvider_Response) (*tfprotov6.ConfigureProviderResponse, error) {
	diags, err := Diagnostics(in.Diagnostics)
	if err != nil {
		return nil, err
	}
	return &tfprotov6.ConfigureProviderResponse{
		Diagnostics: diags,
	}, nil
}

func StopProviderRequest(in *tfplugin6.StopProvider_Request) (*tfprotov6.StopProviderRequest, error) {
	return &tfprotov6.StopProviderRequest{}, nil
}

func StopProviderResponse(in *tfplugin6.StopProvider_Response) (*tfprotov6.StopProviderResponse, error) {
	return &tfprotov6.StopProviderResponse{
		Error: in.Error,
	}, nil
}

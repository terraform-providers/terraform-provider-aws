package plugin

import (
	"context"
	"errors"
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	proto "github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfplugin5"
)

var (
	_ plugin.GRPCPlugin = (*gRPCProviderPlugin)(nil)
	_ plugin.Plugin     = (*gRPCProviderPlugin)(nil)
)

// gRPCProviderPlugin implements plugin.GRPCPlugin and plugin.Plugin for the go-plugin package.
// the only real implementation is GRPCSServer, the other methods are only satisfied
// for compatibility with go-plugin
type gRPCProviderPlugin struct {
	GRPCProvider func() proto.ProviderServer
}

func (p *gRPCProviderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, errors.New("terraform-plugin-sdk only implements grpc servers")
}

func (p *gRPCProviderPlugin) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, errors.New("terraform-plugin-sdk only implements grpc servers")
}

func (p *gRPCProviderPlugin) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (interface{}, error) {
	return nil, errors.New("terraform-plugin-sdk only implements grpc servers")
}

func (p *gRPCProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterProviderServer(s, p.GRPCProvider())
	return nil
}

package plugin

import (
	"context"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	grpcplugin "github.com/hashicorp/terraform-plugin-sdk/v2/internal/helper/plugin"
	proto "github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfplugin5"
)

const (
	// The constants below are the names of the plugins that can be dispensed
	// from the plugin server.
	ProviderPluginName = "provider"
)

// Handshake is the HandshakeConfig used to configure clients and servers.
var Handshake = plugin.HandshakeConfig{
	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

type ProviderFunc func() *schema.Provider
type GRPCProviderFunc func() proto.ProviderServer

// ServeOpts are the configurations to serve a plugin.
type ServeOpts struct {
	ProviderFunc ProviderFunc

	// Wrapped versions of the above plugins will automatically shimmed and
	// added to the GRPC functions when possible.
	GRPCProviderFunc GRPCProviderFunc

	// Logger is the logger that go-plugin will use.
	Logger hclog.Logger

	// TestConfig should only be set when the provider is being tested; it
	// will opt out of go-plugin's lifecycle management and other features,
	// and will use the supplied configuration options to control the
	// plugin's lifecycle and communicate connection information. See the
	// go-plugin GoDoc for more information.
	TestConfig *plugin.ServeTestConfig
}

// Serve serves a plugin. This function never returns and should be the final
// function called in the main function of the plugin.
func Serve(opts *ServeOpts) {
	// since the plugins may not yet be aware of the new protocol, we
	// automatically wrap the plugins in the grpc shims.
	if opts.GRPCProviderFunc == nil && opts.ProviderFunc != nil {
		opts.GRPCProviderFunc = func() proto.ProviderServer {
			return grpcplugin.NewGRPCProviderServer(opts.ProviderFunc())
		}
	}

	provider := opts.GRPCProviderFunc()
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		VersionedPlugins: map[int]plugin.PluginSet{
			5: {
				ProviderPluginName: &gRPCProviderPlugin{
					GRPCProvider: func() proto.ProviderServer {
						return provider
					},
				},
			},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(append(opts, grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				ctx = provider.(*grpcplugin.GRPCProviderServer).StopContext(ctx)
				return handler(ctx, req)
			}))...)
		},
		Logger: opts.Logger,
		Test:   opts.TestConfig,
	})
}

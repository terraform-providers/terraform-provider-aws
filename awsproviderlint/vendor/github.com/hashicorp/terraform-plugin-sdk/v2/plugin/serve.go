package plugin

import (
	"log"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	tf5server "github.com/hashicorp/terraform-plugin-go/tfprotov5/server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
type GRPCProviderFunc func() tfprotov5.ProviderServer

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

	// Set NoLogOutputOverride to not override the log output with an hclog
	// adapter. This should only be used when running the plugin in
	// acceptance tests.
	NoLogOutputOverride bool
}

// Serve serves a plugin. This function never returns and should be the final
// function called in the main function of the plugin.
func Serve(opts *ServeOpts) {
	if !opts.NoLogOutputOverride {
		// In order to allow go-plugin to correctly pass log-levels through to
		// terraform, we need to use an hclog.Logger with JSON output. We can
		// inject this into the std `log` package here, so existing providers will
		// make use of it automatically.
		logger := hclog.New(&hclog.LoggerOptions{
			// We send all output to terraform. Go-plugin will take the output and
			// pass it through another hclog.Logger on the client side where it can
			// be filtered.
			Level:      hclog.Trace,
			JSONFormat: true,
		})
		log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	}

	// since the plugins may not yet be aware of the new protocol, we
	// automatically wrap the plugins in the grpc shims.
	if opts.GRPCProviderFunc == nil && opts.ProviderFunc != nil {
		opts.GRPCProviderFunc = func() tfprotov5.ProviderServer {
			return schema.NewGRPCProviderServer(opts.ProviderFunc())
		}
	}

	provider := opts.GRPCProviderFunc()
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		VersionedPlugins: map[int]plugin.PluginSet{
			5: {
				ProviderPluginName: &tf5server.GRPCProviderPlugin{
					GRPCProvider: func() tfprotov5.ProviderServer {
						return provider
					},
				},
			},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(opts...)
		},
		Logger: opts.Logger,
		Test:   opts.TestConfig,
	})
}

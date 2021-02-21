package tf5server

import (
	"context"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/fromproto"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/toproto"
)

// ServeOpt is an interface for defining options that can be passed to the
// Serve function. Each implementation modifies the ServeConfig being
// generated. A slice of ServeOpts then, cumulatively applied, render a full
// ServeConfig.
type ServeOpt interface {
	ApplyServeOpt(*ServeConfig) error
}

// ServeConfig contains the configured options for how a provider should be
// served.
type ServeConfig struct {
	logger       hclog.Logger
	debugCtx     context.Context
	debugCh      chan *plugin.ReattachConfig
	debugCloseCh chan struct{}
}

type serveConfigFunc func(*ServeConfig) error

func (s serveConfigFunc) ApplyServeOpt(in *ServeConfig) error {
	return s(in)
}

// WithDebug returns a ServeOpt that will set the server into debug mode, using
// the passed options to populate the go-plugin ServeTestConfig.
func WithDebug(ctx context.Context, config chan *plugin.ReattachConfig, closeCh chan struct{}) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.debugCtx = ctx
		in.debugCh = config
		in.debugCloseCh = closeCh
		return nil
	})
}

// WithGoPluginLogger returns a ServeOpt that will set the logger that
// go-plugin should use to log messages.
func WithGoPluginLogger(logger hclog.Logger) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.logger = logger
		return nil
	})
}

// Serve starts a tfprotov5.ProviderServer serving, ready for Terraform to
// connect to it. The name passed in should be the fully qualified name that
// users will enter in the source field of the required_providers block, like
// "registry.terraform.io/hashicorp/time".
//
// Zero or more options to configure the server may also be passed. The default
// invocation is sufficient, but if the provider wants to run in debug mode or
// modify the logger that go-plugin is using, ServeOpts can be specified to
// support that.
func Serve(name string, serverFactory func() tfprotov5.ProviderServer, opts ...ServeOpt) error {
	var conf ServeConfig
	for _, opt := range opts {
		err := opt.ApplyServeOpt(&conf)
		if err != nil {
			return err
		}
	}
	serveConfig := &plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  5,
			MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
		},
		Plugins: plugin.PluginSet{
			"provider": &GRPCProviderPlugin{
				GRPCProvider: serverFactory,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	}
	if conf.logger != nil {
		serveConfig.Logger = conf.logger
	}
	if conf.debugCh != nil {
		serveConfig.Test = &plugin.ServeTestConfig{
			Context:          conf.debugCtx,
			ReattachConfigCh: conf.debugCh,
			CloseCh:          conf.debugCloseCh,
		}
	}
	plugin.Serve(serveConfig)
	return nil
}

type server struct {
	downstream tfprotov5.ProviderServer
	tfplugin5.UnimplementedProviderServer

	stopMu sync.Mutex
	stopCh chan struct{}
}

func mergeStop(ctx context.Context, cancel context.CancelFunc, stopCh chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-stopCh:
		cancel()
	}
}

// stoppableContext returns a context that wraps `ctx` but will be canceled
// when the server's stopCh is closed.
//
// This is used to cancel all in-flight contexts when the Stop method of the
// server is called.
func (s *server) stoppableContext(ctx context.Context) context.Context {
	s.stopMu.Lock()
	defer s.stopMu.Unlock()

	stoppable, cancel := context.WithCancel(ctx)
	go mergeStop(stoppable, cancel, s.stopCh)
	return stoppable
}

// New converts a tfprotov5.ProviderServer into a server capable of handling
// Terraform protocol requests and issuing responses using the gRPC types.
func New(serve tfprotov5.ProviderServer) tfplugin5.ProviderServer {
	return &server{
		downstream: serve,
		stopCh:     make(chan struct{}),
	}
}

func (s *server) GetSchema(ctx context.Context, req *tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.GetProviderSchemaRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.GetProviderSchema(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.GetProviderSchema_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) PrepareProviderConfig(ctx context.Context, req *tfplugin5.PrepareProviderConfig_Request) (*tfplugin5.PrepareProviderConfig_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.PrepareProviderConfigRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.PrepareProviderConfig(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.PrepareProviderConfig_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) Configure(ctx context.Context, req *tfplugin5.Configure_Request) (*tfplugin5.Configure_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.ConfigureProviderRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.ConfigureProvider(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.Configure_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// stop closes the stopCh associated with the server and replaces it with a new
// one.
//
// This causes all in-flight requests for the server to have their contexts
// canceled.
func (s *server) stop() {
	s.stopMu.Lock()
	defer s.stopMu.Unlock()

	close(s.stopCh)
	s.stopCh = make(chan struct{})
}

func (s *server) Stop(ctx context.Context, req *tfplugin5.Stop_Request) (*tfplugin5.Stop_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.StopProviderRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.StopProvider(ctx, r)
	if err != nil {
		return nil, err
	}
	s.stop()
	ret, err := toproto.Stop_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) ValidateDataSourceConfig(ctx context.Context, req *tfplugin5.ValidateDataSourceConfig_Request) (*tfplugin5.ValidateDataSourceConfig_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.ValidateDataSourceConfigRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.ValidateDataSourceConfig(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.ValidateDataSourceConfig_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) ReadDataSource(ctx context.Context, req *tfplugin5.ReadDataSource_Request) (*tfplugin5.ReadDataSource_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.ReadDataSourceRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.ReadDataSource(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.ReadDataSource_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) ValidateResourceTypeConfig(ctx context.Context, req *tfplugin5.ValidateResourceTypeConfig_Request) (*tfplugin5.ValidateResourceTypeConfig_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.ValidateResourceTypeConfigRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.ValidateResourceTypeConfig(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.ValidateResourceTypeConfig_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) UpgradeResourceState(ctx context.Context, req *tfplugin5.UpgradeResourceState_Request) (*tfplugin5.UpgradeResourceState_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.UpgradeResourceStateRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.UpgradeResourceState(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.UpgradeResourceState_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) ReadResource(ctx context.Context, req *tfplugin5.ReadResource_Request) (*tfplugin5.ReadResource_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.ReadResourceRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.ReadResource(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.ReadResource_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) PlanResourceChange(ctx context.Context, req *tfplugin5.PlanResourceChange_Request) (*tfplugin5.PlanResourceChange_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.PlanResourceChangeRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.PlanResourceChange(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.PlanResourceChange_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) ApplyResourceChange(ctx context.Context, req *tfplugin5.ApplyResourceChange_Request) (*tfplugin5.ApplyResourceChange_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.ApplyResourceChangeRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.ApplyResourceChange(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.ApplyResourceChange_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *server) ImportResourceState(ctx context.Context, req *tfplugin5.ImportResourceState_Request) (*tfplugin5.ImportResourceState_Response, error) {
	ctx = s.stoppableContext(ctx)
	r, err := fromproto.ImportResourceStateRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.downstream.ImportResourceState(ctx, r)
	if err != nil {
		return nil, err
	}
	ret, err := toproto.ImportResourceState_Response(resp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

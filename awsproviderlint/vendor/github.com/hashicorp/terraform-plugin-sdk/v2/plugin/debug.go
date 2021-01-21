package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
)

// ReattachConfig holds the information Terraform needs to be able to attach
// itself to a provider process, so it can drive the process.
type ReattachConfig struct {
	Protocol string
	Pid      int
	Test     bool
	Addr     ReattachConfigAddr
}

// ReattachConfigAddr is a JSON-encoding friendly version of net.Addr.
type ReattachConfigAddr struct {
	Network string
	String  string
}

// DebugServe starts a plugin server in debug mode; this should only be used
// when the provider will manage its own lifecycle. It is not recommended for
// normal usage; Serve is the correct function for that.
func DebugServe(ctx context.Context, opts *ServeOpts) (ReattachConfig, <-chan struct{}, error) {
	reattachCh := make(chan *plugin.ReattachConfig)
	closeCh := make(chan struct{})

	opts.TestConfig = &plugin.ServeTestConfig{
		Context:          ctx,
		ReattachConfigCh: reattachCh,
		CloseCh:          closeCh,
	}

	go Serve(opts)

	var config *plugin.ReattachConfig
	select {
	case config = <-reattachCh:
	case <-time.After(2 * time.Second):
		return ReattachConfig{}, closeCh, errors.New("timeout waiting on reattach config")
	}

	if config == nil {
		return ReattachConfig{}, closeCh, errors.New("nil reattach config received")
	}

	return ReattachConfig{
		Protocol: string(config.Protocol),
		Pid:      config.Pid,
		Test:     config.Test,
		Addr: ReattachConfigAddr{
			Network: config.Addr.Network(),
			String:  config.Addr.String(),
		},
	}, closeCh, nil
}

// Debug starts a debug server and controls its lifecycle, printing the
// information needed for Terraform to connect to the provider to stdout.
// os.Interrupt will be captured and used to stop the server.
func Debug(ctx context.Context, providerAddr string, opts *ServeOpts) error {
	ctx, cancel := context.WithCancel(ctx)
	// Ctrl-C will stop the server
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer func() {
		signal.Stop(sigCh)
		cancel()
	}()
	config, closeCh, err := DebugServe(ctx, opts)
	if err != nil {
		return fmt.Errorf("Error launching debug server: %w", err)
	}
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	reattachBytes, err := json.Marshal(map[string]ReattachConfig{
		providerAddr: config,
	})
	if err != nil {
		return fmt.Errorf("Error building reattach string: %w", err)
	}

	reattachStr := string(reattachBytes)

	fmt.Printf("Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:\n\n")
	switch runtime.GOOS {
	case "windows":
		fmt.Printf("\tCommand Prompt:\tset \"TF_REATTACH_PROVIDERS=%s\"\n", reattachStr)
		fmt.Printf("\tPowerShell:\t$env:TF_REATTACH_PROVIDERS='%s'\n", strings.ReplaceAll(reattachStr, `'`, `''`))
	case "linux", "darwin":
		fmt.Printf("\tTF_REATTACH_PROVIDERS='%s'\n", strings.ReplaceAll(reattachStr, `'`, `'"'"'`))
	default:
		fmt.Println(reattachStr)
	}
	fmt.Println("")

	// wait for the server to be done
	<-closeCh
	return nil
}

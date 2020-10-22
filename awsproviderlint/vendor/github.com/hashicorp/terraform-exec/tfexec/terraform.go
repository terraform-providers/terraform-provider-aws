package tfexec

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
)

type printfer interface {
	Printf(format string, v ...interface{})
}

// Terraform represents the Terraform CLI executable and working directory.
//
// Typically this is constructed against the root module of a Terraform configuration
// but you can override paths used in some commands depending on the available
// options.
//
// By default, the instance inherits the environment from the calling code (using os.Environ)
// but it ignores certain environment variables that are managed within the code and prohibits
// setting them through SetEnv:
//
//  - TF_APPEND_USER_AGENT
//  - TF_IN_AUTOMATION
//  - TF_INPUT
//  - TF_LOG
//  - TF_LOG_PATH
//  - TF_REATTACH_PROVIDERS
//  - TF_DISABLE_PLUGIN_TLS
//  - TF_SKIP_PROVIDER_VERIFY
type Terraform struct {
	execPath           string
	workingDir         string
	appendUserAgent    string
	disablePluginTLS   bool
	skipProviderVerify bool
	env                map[string]string

	stdout  io.Writer
	stderr  io.Writer
	logger  printfer
	logPath string

	versionLock  sync.Mutex
	execVersion  *version.Version
	provVersions map[string]*version.Version
}

// NewTerraform returns a Terraform struct with default values for all fields.
// If a blank execPath is supplied, NewTerraform will attempt to locate an
// appropriate binary on the system PATH.
func NewTerraform(workingDir string, execPath string) (*Terraform, error) {
	if workingDir == "" {
		return nil, fmt.Errorf("Terraform cannot be initialised with empty workdir")
	}

	if _, err := os.Stat(workingDir); err != nil {
		return nil, fmt.Errorf("error initialising Terraform with workdir %s: %s", workingDir, err)
	}

	if execPath == "" {
		err := fmt.Errorf("NewTerraform: please supply the path to a Terraform executable using execPath, e.g. using the tfinstall package.")
		return nil, &ErrNoSuitableBinary{err: err}

	}
	tf := Terraform{
		execPath:   execPath,
		workingDir: workingDir,
		env:        nil, // explicit nil means copy os.Environ
		logger:     log.New(ioutil.Discard, "", 0),
	}

	return &tf, nil
}

// SetEnv allows you to override environment variables, this should not be used for any well known
// Terraform environment variables that are already covered in options. Pass nil to copy the values
// from os.Environ. Attempting to set environment variables that should be managed manually will
// result in ErrManualEnvVar being returned.
func (tf *Terraform) SetEnv(env map[string]string) error {
	for k := range env {
		if strings.HasPrefix(k, varEnvVarPrefix) {
			return fmt.Errorf("variables should be passed using the Var option: %w", &ErrManualEnvVar{k})
		}
		for _, p := range prohibitedEnvVars {
			if p == k {
				return &ErrManualEnvVar{k}
			}
		}
	}

	tf.env = env
	return nil
}

// SetLogger specifies a logger for tfexec to use.
func (tf *Terraform) SetLogger(logger printfer) {
	tf.logger = logger
}

// SetStdout specifies a writer to stream stdout to for every command.
//
// This should be used for information or logging purposes only, not control
// flow. Any parsing necessary should be added as functionality to this package.
func (tf *Terraform) SetStdout(w io.Writer) {
	tf.stdout = w
}

// SetStderr specifies a writer to stream stderr to for every command.
//
// This should be used for information or logging purposes only, not control
// flow. Any parsing necessary should be added as functionality to this package.
func (tf *Terraform) SetStderr(w io.Writer) {
	tf.stderr = w
}

// SetLogPath sets the TF_LOG_PATH environment variable for Terraform CLI
// execution.
func (tf *Terraform) SetLogPath(path string) error {
	tf.logPath = path
	return nil
}

// SetAppendUserAgent sets the TF_APPEND_USER_AGENT environment variable for
// Terraform CLI execution.
func (tf *Terraform) SetAppendUserAgent(ua string) error {
	tf.appendUserAgent = ua
	return nil
}

// SetDisablePluginTLS sets the TF_DISABLE_PLUGIN_TLS environment variable for
// Terraform CLI execution.
func (tf *Terraform) SetDisablePluginTLS(disabled bool) error {
	tf.disablePluginTLS = disabled
	return nil
}

// SetSkipProviderVerify sets the TF_SKIP_PROVIDER_VERIFY environment variable
// for Terraform CLI execution. This is no longer used in 0.13.0 and greater.
func (tf *Terraform) SetSkipProviderVerify(skip bool) error {
	err := tf.compatible(context.Background(), nil, tf0_13_0)
	if err != nil {
		return err
	}
	tf.skipProviderVerify = skip
	return nil
}

// WorkingDir returns the working directory for Terraform.
func (tf *Terraform) WorkingDir() string {
	return tf.workingDir
}

// ExecPath returns the path to the Terraform executable.
func (tf *Terraform) ExecPath() string {
	return tf.execPath
}

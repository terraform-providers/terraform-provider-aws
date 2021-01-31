package plugintest

import (
	"fmt"
	"io/ioutil"
	"os"
)

const subprocessCurrentSigil = "4acd63807899403ca4859f5bb948d2c6"
const subprocessPreviousSigil = "2279afb8cf71423996be1fd65d32f13b"

// AutoInitProviderHelper is the main entrypoint for testing provider plugins
// using this package. It is intended to be called during TestMain to prepare
// for provider testing.
//
// AutoInitProviderHelper will discover the location of a current Terraform CLI
// executable to test against, detect whether a prior version of the plugin is
// available for upgrade tests, and then will return an object containing the
// results of that initialization which can then be stored in a global variable
// for use in other tests.
func AutoInitProviderHelper(sourceDir string) *Helper {
	helper, err := AutoInitHelper(sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot run Terraform provider tests: %s\n", err)
		os.Exit(1)
	}
	return helper
}

// Helper is intended as a per-package singleton created in TestMain which
// other tests in a package can use to create Terraform execution contexts
type Helper struct {
	baseDir string

	// sourceDir is the dir containing the provider source code, needed
	// for tests that use fixture files.
	sourceDir     string
	terraformExec string

	// execTempDir is created during DiscoverConfig to store any downloaded
	// binaries
	execTempDir string
}

// AutoInitHelper uses the auto-discovery behavior of DiscoverConfig to prepare
// a configuration and then calls InitHelper with it. This is a convenient
// way to get the standard init behavior based on environment variables, and
// callers should use this unless they have an unusual requirement that calls
// for constructing a config in a different way.
func AutoInitHelper(sourceDir string) (*Helper, error) {
	config, err := DiscoverConfig(sourceDir)
	if err != nil {
		return nil, err
	}

	return InitHelper(config)
}

// InitHelper prepares a testing helper with the given configuration.
//
// For most callers it is sufficient to call AutoInitHelper instead, which
// will construct a configuration automatically based on certain environment
// variables.
//
// If this function returns an error then it may have left some temporary files
// behind in the system's temporary directory. There is currently no way to
// automatically clean those up.
func InitHelper(config *Config) (*Helper, error) {
	tempDir := os.Getenv("TF_ACC_TEMP_DIR")
	baseDir, err := ioutil.TempDir(tempDir, "plugintest")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory for test helper: %s", err)
	}

	return &Helper{
		baseDir:       baseDir,
		sourceDir:     config.SourceDir,
		terraformExec: config.TerraformExec,
		execTempDir:   config.execTempDir,
	}, nil
}

// Close cleans up temporary files and directories created to support this
// helper, returning an error if any of the cleanup fails.
//
// Call this before returning from TestMain to minimize the amount of detritus
// left behind in the filesystem after the tests complete.
func (h *Helper) Close() error {
	if h.execTempDir != "" {
		err := os.RemoveAll(h.execTempDir)
		if err != nil {
			return err
		}
	}
	return os.RemoveAll(h.baseDir)
}

// NewWorkingDir creates a new working directory for use in the implementation
// of a single test, returning a WorkingDir object representing that directory.
//
// If the working directory object is not itself closed by the time the test
// program exits, the Close method on the helper itself will attempt to
// delete it.
func (h *Helper) NewWorkingDir() (*WorkingDir, error) {
	dir, err := ioutil.TempDir(h.baseDir, "work")
	if err != nil {
		return nil, err
	}

	// symlink the provider source files into the base directory
	err = symlinkDirectoriesOnly(h.sourceDir, dir)
	if err != nil {
		return nil, err
	}

	return &WorkingDir{
		h:             h,
		baseDir:       dir,
		terraformExec: h.terraformExec,
	}, nil
}

// RequireNewWorkingDir is a variant of NewWorkingDir that takes a TestControl
// object and will immediately fail the running test if the creation of the
// working directory fails.
func (h *Helper) RequireNewWorkingDir(t TestControl) *WorkingDir {
	t.Helper()

	wd, err := h.NewWorkingDir()
	if err != nil {
		t := testingT{t}
		t.Fatalf("failed to create new working directory: %s", err)
		return nil
	}
	return wd
}

// TerraformExecPath returns the location of the Terraform CLI executable that
// should be used when running tests.
func (h *Helper) TerraformExecPath() string {
	return h.terraformExec
}

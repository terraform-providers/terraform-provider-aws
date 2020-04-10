package tftest

import (
	"bytes"
	"encoding/json"
	"fmt"
	getter "github.com/hashicorp/go-getter"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const releaseHost = "https://releases.hashicorp.com"

// FindTerraform attempts to find a Terraform CLI executable for plugin testing.
//
// As a first preference it will look for the environment variable
// TF_ACC_TERRAFORM_PATH and return its value. If that variable is not set, it will
// look in PATH for a program named "terraform" and, if one is found, return
// its absolute path.
//
// If no Terraform executable can be found, the result is the empty string. In
// that case, the test program will usually fail outright.
func FindTerraform() string {
	if p := os.Getenv("TF_ACC_TERRAFORM_PATH"); p != "" {
		return p
	}
	p, err := exec.LookPath("terraform")
	if err != nil {
		return ""
	}
	return p
}

func tfURL(version, osName, archName string) string {
	return fmt.Sprintf(
		"%s/terraform/%s/terraform_%s_%s_%s.zip",
		releaseHost, version, version, osName, archName,
	)
}

// InstallTerraform downloads and decompresses a Terraform CLI executable with
// the specified version, downloaded from the HashiCorp releases page over HTTP.
//
// The version string must match an existing Terraform release semver version,
// e.g. 0.12.5.
//
// The terraform executable is installed to a temporary folder.
//
// FIXME: Temporary folder should be cleaned up after tests have finished.
func InstallTerraform(tfVersion string) (string, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	var tfDir string
	var err error

	tempDir := os.Getenv("TF_ACC_TEMP_DIR")
	tfDir, err = ioutil.TempDir(tempDir, "tftest-terraform")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %s", err)
	}

	url := tfURL(tfVersion, osName, archName)

	client := getter.Client{
		Src: url,
		Dst: tfDir,

		Mode: getter.ClientModeDir,
	}

	err = client.Get()
	if err != nil {
		return "", fmt.Errorf("failed to download terraform from %s: %s", url, err)
	}

	return filepath.Join(tfDir, "terraform"), nil
}

// getTerraformEnv returns the appropriate Env for the Terraform command.
func getTerraformEnv() []string {
	var env []string
	for _, e := range os.Environ() {
		env = append(env, e)
	}

	env = append(env, "TF_DISABLE_PLUGIN_TLS=1")
	env = append(env, "TF_SKIP_PROVIDER_VERIFY=1")

	// FIXME: Ideally in testing.Verbose mode we'd turn on Terraform DEBUG
	// logging, perhaps redirected to a separate fd other than stderr to avoid
	// polluting it, and then propagate the log lines out into t.Log so that
	// they are visible to the person running the test. Currently though,
	// Terraform CLI is able to send logs only to either an on-disk file or
	// to stderr.
	env = append(env, "TF_LOG=") // so logging can't pollute our stderr output
	env = append(env, "TF_INPUT=0")

	if p := os.Getenv("TF_ACC_LOG_PATH"); p != "" {
		env = append(env, "TF_LOG=TRACE")
		env = append(env, "TF_LOG_PATH="+p)
	}
	return env
}

// RunTerraform runs the configured Terraform CLI executable with the given
// arguments, returning an error if it produces a non-successful exit status.
func (wd *WorkingDir) runTerraform(args ...string) error {
	allArgs := []string{"terraform"}
	allArgs = append(allArgs, args...)

	env := getTerraformEnv()

	var errBuf strings.Builder

	cmd := &exec.Cmd{
		Path:   wd.h.TerraformExecPath(),
		Args:   allArgs,
		Dir:    wd.baseDir,
		Stderr: &errBuf,
		Env:    env,
	}
	err := cmd.Run()
	if tErr, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("terraform failed: %s\n\nstderr:\n%s", tErr.ProcessState.String(), errBuf.String())
	}
	return err
}

// runTerraformJSON runs the configured Terraform CLI executable with the given
// arguments and tries to decode its stdout into the given target value (which
// must be a non-nil pointer) as JSON.
func (wd *WorkingDir) runTerraformJSON(target interface{}, args ...string) error {
	allArgs := []string{"terraform"}
	allArgs = append(allArgs, args...)

	env := getTerraformEnv()

	var outBuf bytes.Buffer
	var errBuf strings.Builder

	cmd := &exec.Cmd{
		Path:   wd.h.TerraformExecPath(),
		Args:   allArgs,
		Dir:    wd.baseDir,
		Stderr: &errBuf,
		Stdout: &outBuf,
		Env:    env,
	}
	err := cmd.Run()
	if err != nil {
		if tErr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("terraform failed: %s\n\nstderr:\n%s", tErr.ProcessState.String(), errBuf.String())
		}
		return err
	}

	return json.Unmarshal(outBuf.Bytes(), target)
}

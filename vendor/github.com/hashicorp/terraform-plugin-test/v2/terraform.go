package tftest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// getTerraformEnv returns the appropriate Env for the Terraform command.
func (wd *WorkingDir) getTerraformEnv() []string {
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

	// don't propagate the magic cookie
	env = append(env, "TF_PLUGIN_MAGIC_COOKIE=")

	if p := os.Getenv("TF_ACC_LOG_PATH"); p != "" {
		env = append(env, "TF_LOG=TRACE")
		env = append(env, "TF_LOG_PATH="+p)
	}

	for k, v := range wd.env {
		env = append(env, k+"="+v)
	}
	return env
}

// runTerraform runs the configured Terraform CLI executable with the given
// arguments, returning an error if it produces a non-successful exit status.
// if captureStdout is non-nil, the process will write it's stdout to the
// provided io.Writer
func (wd *WorkingDir) runTerraform(captureStdout io.Writer, args ...string) error {
	allArgs := []string{"terraform"}
	allArgs = append(allArgs, args...)

	env := wd.getTerraformEnv()

	var errBuf strings.Builder

	cmd := &exec.Cmd{
		Path:   wd.h.TerraformExecPath(),
		Args:   allArgs,
		Dir:    wd.baseDir,
		Stderr: &errBuf,
		Env:    env,
	}

	if captureStdout != nil {
		cmd.Stdout = captureStdout
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

	env := wd.getTerraformEnv()

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

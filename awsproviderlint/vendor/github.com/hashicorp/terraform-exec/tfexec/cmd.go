package tfexec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-exec/internal/version"
)

const (
	checkpointDisableEnvVar  = "CHECKPOINT_DISABLE"
	cliArgsEnvVar            = "TF_CLI_ARGS"
	logEnvVar                = "TF_LOG"
	inputEnvVar              = "TF_INPUT"
	automationEnvVar         = "TF_IN_AUTOMATION"
	logPathEnvVar            = "TF_LOG_PATH"
	reattachEnvVar           = "TF_REATTACH_PROVIDERS"
	appendUserAgentEnvVar    = "TF_APPEND_USER_AGENT"
	workspaceEnvVar          = "TF_WORKSPACE"
	disablePluginTLSEnvVar   = "TF_DISABLE_PLUGIN_TLS"
	skipProviderVerifyEnvVar = "TF_SKIP_PROVIDER_VERIFY"

	varEnvVarPrefix    = "TF_VAR_"
	cliArgEnvVarPrefix = "TF_CLI_ARGS_"
)

var prohibitedEnvVars = []string{
	cliArgsEnvVar,
	inputEnvVar,
	automationEnvVar,
	logPathEnvVar,
	logEnvVar,
	reattachEnvVar,
	appendUserAgentEnvVar,
	workspaceEnvVar,
	disablePluginTLSEnvVar,
	skipProviderVerifyEnvVar,
}

var prohibitedEnvVarPrefixes = []string{
	varEnvVarPrefix,
	cliArgEnvVarPrefix,
}

func manualEnvVars(env map[string]string, cb func(k string)) {
	for k := range env {
		for _, p := range prohibitedEnvVars {
			if p == k {
				cb(k)
				goto NextEnvVar
			}
		}
		for _, prefix := range prohibitedEnvVarPrefixes {
			if strings.HasPrefix(k, prefix) {
				cb(k)
				goto NextEnvVar
			}
		}
	NextEnvVar:
	}
}

// ProhibitedEnv returns a slice of environment variable keys that are not allowed
// to be set manually from the passed environment.
func ProhibitedEnv(env map[string]string) []string {
	var p []string
	manualEnvVars(env, func(k string) {
		p = append(p, k)
	})
	return p
}

// CleanEnv removes any prohibited environment variables from an environment map.
func CleanEnv(dirty map[string]string) map[string]string {
	clean := dirty
	manualEnvVars(clean, func(k string) {
		delete(clean, k)
	})
	return clean
}

func envMap(environ []string) map[string]string {
	env := map[string]string{}
	for _, ev := range environ {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) == 0 {
			continue
		}
		k := parts[0]
		v := ""
		if len(parts) == 2 {
			v = parts[1]
		}
		env[k] = v
	}
	return env
}

func envSlice(environ map[string]string) []string {
	env := []string{}
	for k, v := range environ {
		env = append(env, k+"="+v)
	}
	return env
}

func (tf *Terraform) buildEnv(mergeEnv map[string]string) []string {
	// set Terraform level env, if env is nil, fall back to os.Environ
	var env map[string]string
	if tf.env == nil {
		env = envMap(os.Environ())
	} else {
		env = make(map[string]string, len(tf.env))
		for k, v := range tf.env {
			env[k] = v
		}
	}

	// override env with any command specific environment
	for k, v := range mergeEnv {
		env[k] = v
	}

	// always propagate CHECKPOINT_DISABLE env var unless it is
	// explicitly overridden with tf.SetEnv or command env
	if _, ok := env[checkpointDisableEnvVar]; !ok {
		env[checkpointDisableEnvVar] = os.Getenv(checkpointDisableEnvVar)
	}

	// always override user agent
	ua := mergeUserAgent(
		os.Getenv(appendUserAgentEnvVar),
		tf.appendUserAgent,
		fmt.Sprintf("HashiCorp-terraform-exec/%s", version.ModuleVersion()),
	)
	env[appendUserAgentEnvVar] = ua

	// always override logging
	if tf.logPath == "" {
		// so logging can't pollute our stderr output
		env[logEnvVar] = ""
		env[logPathEnvVar] = ""
	} else {
		env[logPathEnvVar] = tf.logPath
		// Log levels other than TRACE are currently unreliable, the CLI recommends using TRACE only.
		env[logEnvVar] = "TRACE"
	}

	// constant automation override env vars
	env[automationEnvVar] = "1"

	// force usage of workspace methods for switching
	env[workspaceEnvVar] = ""

	if tf.disablePluginTLS {
		env[disablePluginTLSEnvVar] = "1"
	}

	if tf.skipProviderVerify {
		env[skipProviderVerifyEnvVar] = "1"
	}

	return envSlice(env)
}

func (tf *Terraform) buildTerraformCmd(ctx context.Context, mergeEnv map[string]string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, tf.execPath, args...)
	cmd.Env = tf.buildEnv(mergeEnv)
	cmd.Dir = tf.workingDir

	tf.logger.Printf("[INFO] running Terraform command: %s", cmdString(cmd))

	return cmd
}

func (tf *Terraform) runTerraformCmdJSON(cmd *exec.Cmd, v interface{}) error {
	var outbuf = bytes.Buffer{}
	cmd.Stdout = mergeWriters(cmd.Stdout, &outbuf)

	err := tf.runTerraformCmd(cmd)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(&outbuf)
	dec.UseNumber()
	return dec.Decode(v)
}

func (tf *Terraform) runTerraformCmd(cmd *exec.Cmd) error {
	var errBuf strings.Builder

	cmd.Stdout = mergeWriters(cmd.Stdout, tf.stdout)
	cmd.Stderr = mergeWriters(cmd.Stderr, tf.stderr, &errBuf)

	err := cmd.Run()
	if err != nil {
		return tf.parseError(err, errBuf.String())
	}
	return nil
}

// mergeUserAgent does some minor deduplication to ensure we aren't
// just using the same append string over and over.
func mergeUserAgent(uas ...string) string {
	included := map[string]bool{}
	merged := []string{}
	for _, ua := range uas {
		ua = strings.TrimSpace(ua)

		if ua == "" {
			continue
		}
		if included[ua] {
			continue
		}
		included[ua] = true
		merged = append(merged, ua)
	}
	return strings.Join(merged, " ")
}

func mergeWriters(writers ...io.Writer) io.Writer {
	compact := []io.Writer{}
	for _, w := range writers {
		if w != nil {
			compact = append(compact, w)
		}
	}
	if len(compact) == 0 {
		return ioutil.Discard
	}
	if len(compact) == 1 {
		return compact[0]
	}
	return io.MultiWriter(compact...)
}

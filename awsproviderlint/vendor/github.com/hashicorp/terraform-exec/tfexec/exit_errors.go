package tfexec

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// this file contains errors parsed from stderr

var (
	// The "Required variable not set:" case is for 0.11
	missingVarErrRegexp  = regexp.MustCompile(`Error: No value for required variable|Error: Required variable not set:`)
	missingVarNameRegexp = regexp.MustCompile(`The root module input variable\s"(.+)"\sis\snot\sset,\sand\shas\sno\sdefault|Error: Required variable not set: (.+)`)

	usageRegexp = regexp.MustCompile(`Too many command line arguments|^Usage: .*Options:.*|Error: Invalid -\d+ option`)

	// "Could not load plugin" is present in 0.13
	noInitErrRegexp = regexp.MustCompile(`Error: Could not satisfy plugin requirements|Error: Could not load plugin`)

	noConfigErrRegexp = regexp.MustCompile(`Error: No configuration files`)

	workspaceDoesNotExistRegexp = regexp.MustCompile(`Workspace "(.+)" doesn't exist.`)

	workspaceAlreadyExistsRegexp = regexp.MustCompile(`Workspace "(.+)" already exists`)

	tfVersionMismatchErrRegexp        = regexp.MustCompile(`Error: The currently running version of Terraform doesn't meet the|Error: Unsupported Terraform Core version`)
	tfVersionMismatchConstraintRegexp = regexp.MustCompile(`required_version = "(.+)"|Required version: (.+)\b`)
	configInvalidErrRegexp            = regexp.MustCompile(`There are some problems with the configuration, described below.`)
)

func (tf *Terraform) wrapExitError(ctx context.Context, err error, stderr string) error {
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		// not an exit error, short circuit, nothing to wrap
		return err
	}

	ctxErr := ctx.Err()

	// nothing to parse, return early
	errString := strings.TrimSpace(stderr)
	if errString == "" {
		return &unwrapper{exitErr, ctxErr}
	}

	switch {
	case tfVersionMismatchErrRegexp.MatchString(stderr):
		constraint := ""
		constraints := tfVersionMismatchConstraintRegexp.FindStringSubmatch(stderr)
		for i := 1; i < len(constraints); i++ {
			constraint = strings.TrimSpace(constraints[i])
			if constraint != "" {
				break
			}
		}

		if constraint == "" {
			// hardcode a value here for weird cases (incl. 0.12)
			constraint = "unknown"
		}

		// only set this if it happened to be cached already
		ver := ""
		if tf != nil && tf.execVersion != nil {
			ver = tf.execVersion.String()
		}

		return &ErrTFVersionMismatch{
			unwrapper: unwrapper{exitErr, ctxErr},

			Constraint: constraint,
			TFVersion:  ver,
		}
	case missingVarErrRegexp.MatchString(stderr):
		name := ""
		names := missingVarNameRegexp.FindStringSubmatch(stderr)
		for i := 1; i < len(names); i++ {
			name = strings.TrimSpace(names[i])
			if name != "" {
				break
			}
		}

		return &ErrMissingVar{
			unwrapper: unwrapper{exitErr, ctxErr},

			VariableName: name,
		}
	case usageRegexp.MatchString(stderr):
		return &ErrCLIUsage{
			unwrapper: unwrapper{exitErr, ctxErr},

			stderr: stderr,
		}
	case noInitErrRegexp.MatchString(stderr):
		return &ErrNoInit{
			unwrapper: unwrapper{exitErr, ctxErr},

			stderr: stderr,
		}
	case noConfigErrRegexp.MatchString(stderr):
		return &ErrNoConfig{
			unwrapper: unwrapper{exitErr, ctxErr},

			stderr: stderr,
		}
	case workspaceDoesNotExistRegexp.MatchString(stderr):
		submatches := workspaceDoesNotExistRegexp.FindStringSubmatch(stderr)
		if len(submatches) == 2 {
			return &ErrNoWorkspace{
				unwrapper: unwrapper{exitErr, ctxErr},

				Name: submatches[1],
			}
		}
	case workspaceAlreadyExistsRegexp.MatchString(stderr):
		submatches := workspaceAlreadyExistsRegexp.FindStringSubmatch(stderr)
		if len(submatches) == 2 {
			return &ErrWorkspaceExists{
				unwrapper: unwrapper{exitErr, ctxErr},

				Name: submatches[1],
			}
		}
	case configInvalidErrRegexp.MatchString(stderr):
		return &ErrConfigInvalid{stderr: stderr}
	}

	return fmt.Errorf("%w\n%s", &unwrapper{exitErr, ctxErr}, stderr)
}

type unwrapper struct {
	err    error
	ctxErr error
}

func (u *unwrapper) Unwrap() error {
	return u.err
}

func (u *unwrapper) Is(target error) bool {
	switch target {
	case context.DeadlineExceeded, context.Canceled:
		return u.ctxErr == context.DeadlineExceeded ||
			u.ctxErr == context.Canceled
	}
	return false
}

func (u *unwrapper) Error() string {
	return u.err.Error()
}

type ErrConfigInvalid struct {
	stderr string
}

func (e *ErrConfigInvalid) Error() string {
	return "configuration is invalid"
}

type ErrMissingVar struct {
	unwrapper

	VariableName string
}

func (err *ErrMissingVar) Error() string {
	return fmt.Sprintf("variable %q was required but not supplied", err.VariableName)
}

type ErrNoWorkspace struct {
	unwrapper

	Name string
}

func (err *ErrNoWorkspace) Error() string {
	return fmt.Sprintf("workspace %q does not exist", err.Name)
}

// ErrWorkspaceExists is returned when creating a workspace that already exists
type ErrWorkspaceExists struct {
	unwrapper

	Name string
}

func (err *ErrWorkspaceExists) Error() string {
	return fmt.Sprintf("workspace %q already exists", err.Name)
}

type ErrNoInit struct {
	unwrapper

	stderr string
}

func (e *ErrNoInit) Error() string {
	return e.stderr
}

type ErrNoConfig struct {
	unwrapper

	stderr string
}

func (e *ErrNoConfig) Error() string {
	return e.stderr
}

// ErrCLIUsage is returned when the combination of flags or arguments is incorrect.
//
//  CLI indicates usage errors in three different ways: either
// 1. Exit 1, with a custom error message on stderr.
// 2. Exit 1, with command usage logged to stderr.
// 3. Exit 127, with command usage logged to stdout.
// Currently cases 1 and 2 are handled.
// TODO KEM: Handle exit 127 case. How does this work on non-Unix platforms?
type ErrCLIUsage struct {
	unwrapper

	stderr string
}

func (e *ErrCLIUsage) Error() string {
	return e.stderr
}

// ErrTFVersionMismatch is returned when the running Terraform version is not compatible with the
// value specified for required_version in the terraform block.
type ErrTFVersionMismatch struct {
	unwrapper

	TFVersion string

	// Constraint is not returned in the error messaging on 0.12
	Constraint string
}

func (e *ErrTFVersionMismatch) Error() string {
	return "terraform core version not supported by configuration"
}

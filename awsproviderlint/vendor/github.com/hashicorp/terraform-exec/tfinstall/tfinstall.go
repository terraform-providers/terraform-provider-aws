package tfinstall

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

const baseURL = "https://releases.hashicorp.com/terraform"

type ExecPathFinder interface {
	ExecPath(context.Context) (string, error)
}

func Find(ctx context.Context, opts ...ExecPathFinder) (string, error) {
	var terraformPath string

	// go through the options in order
	// until a valid terraform executable is found
	for _, opt := range opts {
		p, err := opt.ExecPath(ctx)
		if err != nil {
			return "", fmt.Errorf("unexpected error: %s", err)
		}

		if p == "" {
			// strategy did not locate an executable - fall through to next
			continue
		} else {
			terraformPath = p
			break
		}
	}

	if terraformPath == "" {
		return "", fmt.Errorf("could not find terraform executable")
	}

	err := runTerraformVersion(terraformPath)
	if err != nil {
		return "", fmt.Errorf("executable found at path %s is not terraform: %s", terraformPath, err)
	}

	return terraformPath, nil
}

func runTerraformVersion(execPath string) error {
	cmd := exec.Command(execPath, "version")

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	// very basic sanity check
	if !strings.Contains(string(out), "Terraform v") {
		return fmt.Errorf("located executable at %s, but output of `terraform version` was:\n%s", execPath, out)
	}

	return nil
}

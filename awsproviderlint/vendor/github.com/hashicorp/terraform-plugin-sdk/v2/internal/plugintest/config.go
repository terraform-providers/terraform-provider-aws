package plugintest

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/terraform-exec/tfinstall"
)

// Config is used to configure the test helper. In most normal test programs
// the configuration is discovered automatically by an Init* function using
// DiscoverConfig, but this is exposed so that more complex scenarios can be
// implemented by direct configuration.
type Config struct {
	SourceDir          string
	TerraformExec      string
	execTempDir        string
	PreviousPluginExec string
}

// DiscoverConfig uses environment variables and other means to automatically
// discover a reasonable test helper configuration.
func DiscoverConfig(sourceDir string) (*Config, error) {
	tfVersion := strings.TrimPrefix(os.Getenv("TF_ACC_TERRAFORM_VERSION"), "v")
	tfPath := os.Getenv("TF_ACC_TERRAFORM_PATH")

	tempDir := os.Getenv("TF_ACC_TEMP_DIR")
	tfDir, err := ioutil.TempDir(tempDir, "plugintest-terraform")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	finders := []tfinstall.ExecPathFinder{}
	switch {
	case tfPath != "":
		finders = append(finders, tfinstall.ExactPath(tfPath))
	case tfVersion != "":
		finders = append(finders, tfinstall.ExactVersion(tfVersion, tfDir))
	default:
		finders = append(finders, tfinstall.LookPath(), tfinstall.LatestVersion(tfDir, true))
	}
	tfExec, err := tfinstall.Find(context.Background(), finders...)
	if err != nil {
		return nil, err
	}

	return &Config{
		SourceDir:     sourceDir,
		TerraformExec: tfExec,
		execTempDir:   tfDir,
	}, nil
}

package tfinstall

import (
	"context"

	"github.com/hashicorp/go-version"
)

type ExactVersionOption struct {
	tfVersion  string
	installDir string

	UserAgent string
}

var _ ExecPathFinder = &ExactVersionOption{}

func ExactVersion(tfVersion string, installDir string) *ExactVersionOption {
	opt := &ExactVersionOption{
		tfVersion:  tfVersion,
		installDir: installDir,
	}

	return opt
}

func (opt *ExactVersionOption) ExecPath(ctx context.Context) (string, error) {
	// validate version
	_, err := version.NewVersion(opt.tfVersion)
	if err != nil {
		return "", err
	}

	return downloadWithVerification(ctx, opt.tfVersion, opt.installDir, opt.UserAgent)
}

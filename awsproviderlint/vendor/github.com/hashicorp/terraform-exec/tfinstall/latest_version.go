package tfinstall

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-checkpoint"
)

type LatestVersionOption struct {
	forceCheckpoint bool
	installDir      string

	UserAgent string
}

var _ ExecPathFinder = &LatestVersionOption{}

func LatestVersion(installDir string, forceCheckpoint bool) *LatestVersionOption {
	opt := &LatestVersionOption{
		forceCheckpoint: forceCheckpoint,
		installDir:      installDir,
	}

	return opt
}

func (opt *LatestVersionOption) ExecPath(ctx context.Context) (string, error) {
	v, err := latestVersion(opt.forceCheckpoint)
	if err != nil {
		return "", err
	}

	return downloadWithVerification(ctx, v, opt.installDir, opt.UserAgent)
}

func latestVersion(forceCheckpoint bool) (string, error) {
	resp, err := checkpoint.Check(&checkpoint.CheckParams{
		Product: "terraform",
		Force:   forceCheckpoint,
	})
	if err != nil {
		return "", err
	}

	if resp.CurrentVersion == "" {
		return "", fmt.Errorf("could not determine latest version of terraform using checkpoint: CHECKPOINT_DISABLE may be set")
	}

	return resp.CurrentVersion, nil
}

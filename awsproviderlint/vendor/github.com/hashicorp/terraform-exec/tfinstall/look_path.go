package tfinstall

import (
	"context"
	"log"
	"os/exec"
)

type LookPathOption struct {
}

var _ ExecPathFinder = &LookPathOption{}

func LookPath() *LookPathOption {
	opt := &LookPathOption{}

	return opt
}

func (opt *LookPathOption) ExecPath(context.Context) (string, error) {
	p, err := exec.LookPath("terraform")
	if err != nil {
		if notFoundErr, ok := err.(*exec.Error); ok && notFoundErr.Err == exec.ErrNotFound {
			log.Printf("[WARN] could not locate a terraform executable on system path; continuing")
			return "", nil
		}
		return "", err
	}
	return p, nil
}

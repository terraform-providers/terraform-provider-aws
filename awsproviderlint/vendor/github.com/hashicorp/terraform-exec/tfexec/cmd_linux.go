package tfexec

import (
	"context"
	"os/exec"
	"strings"
	"syscall"
)

func (tf *Terraform) runTerraformCmd(ctx context.Context, cmd *exec.Cmd) error {
	var errBuf strings.Builder

	cmd.Stdout = mergeWriters(cmd.Stdout, tf.stdout)
	cmd.Stderr = mergeWriters(cmd.Stderr, tf.stderr, &errBuf)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		// kill children if parent is dead
		Pdeathsig: syscall.SIGKILL,
		// set process group ID
		Setpgid: true,
	}

	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			if cmd != nil && cmd.Process != nil && cmd.ProcessState != nil {
				// send SIGINT to process group
				err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
				if err != nil {
					tf.logger.Printf("error from SIGINT: %s", err)
				}
			}

			// TODO: send a kill if it doesn't respond for a bit?
		}
	}()

	// check for early cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := cmd.Run()
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}
	if err != nil {
		return tf.wrapExitError(ctx, err, errBuf.String())
	}

	return nil
}

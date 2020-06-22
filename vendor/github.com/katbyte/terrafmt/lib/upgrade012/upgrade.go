package upgrade012

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/katbyte/terrafmt/lib/common"
)

func Block(b string) (string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// Make temp directory
	dir, err := ioutil.TempDir(".", "tmp-module")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	// Create temp file
	tmpFile, err := ioutil.TempFile(dir, "*.tf")
	if err != nil {
		return "", err
	}

	defer os.Remove(tmpFile.Name()) // clean up

	// Write from Reader to File
	if _, err := tmpFile.Write(bytes.NewBufferString(b).Bytes()); err != nil {
		tmpFile.Close()
		log.Fatal(err)
	}

	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("terraform", "init", dir)
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("cmd.Run() failed in terraform init with %s: %s", err, stderr)
	}

	defer os.RemoveAll(".terraform") // clean up

	common.Log.Debugf("running terraform... ")
	cmd = exec.Command("terraform", "0.12upgrade", "-yes", dir)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()

	if err != nil {
		_, err := fmt.Println(stdout)
		if err != nil {
			return "", fmt.Errorf("cmd.Run() failed in terraform 0.12upgrade with %s: %s | %s", err, stdout, stderr)
		}

		return "", fmt.Errorf("cmd.Run() failed in terraform 0.12upgrade with %s: %s", err, stderr)
	}

	ec := cmd.ProcessState.ExitCode()
	common.Log.Debugf("terraform exited with %d", ec)
	if ec != 0 {
		return "", fmt.Errorf("terraform failed with %d: %s", ec, stderr)
	}

	// Read from temp file
	raw, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("terrafmt failed with readfile: %s", err)
	}

	// 0.12upgrade always adds a trailing newline, even if it's already there
	// strip it here
	fb := string(raw)
	if strings.HasSuffix(fb, "\n") {
		fb = strings.TrimSuffix(fb, "\n")
	}

	return fb, nil
}

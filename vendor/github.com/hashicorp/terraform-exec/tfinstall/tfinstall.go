package tfinstall

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-checkpoint"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/go-version"
	"golang.org/x/crypto/openpgp"
)

const baseUrl = "https://releases.hashicorp.com/terraform"

type ExecPathFinder interface {
	ExecPath() (string, error)
}

type ExactPathOption struct {
	execPath string
}

func ExactPath(execPath string) *ExactPathOption {
	opt := &ExactPathOption{
		execPath: execPath,
	}
	return opt
}

func (opt *ExactPathOption) ExecPath() (string, error) {
	if _, err := os.Stat(opt.execPath); err != nil {
		// fall through to the next strategy if the local path does not exist
		return "", nil
	}
	return opt.execPath, nil
}

type LookPathOption struct {
}

func LookPath() *LookPathOption {
	opt := &LookPathOption{}

	return opt
}

func (opt *LookPathOption) ExecPath() (string, error) {
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

type LatestVersionOption struct {
	forceCheckpoint bool
	installDir      string
}

func LatestVersion(installDir string, forceCheckpoint bool) *LatestVersionOption {
	opt := &LatestVersionOption{
		forceCheckpoint: forceCheckpoint,
		installDir:      installDir,
	}

	return opt
}

func (opt *LatestVersionOption) ExecPath() (string, error) {
	v, err := latestVersion(opt.forceCheckpoint)
	if err != nil {
		return "", err
	}

	return downloadWithVerification(v, opt.installDir)
}

type ExactVersionOption struct {
	tfVersion  string
	installDir string
}

func ExactVersion(tfVersion string, installDir string) *ExactVersionOption {
	opt := &ExactVersionOption{
		tfVersion:  tfVersion,
		installDir: installDir,
	}

	return opt
}

func (opt *ExactVersionOption) ExecPath() (string, error) {
	// validate version
	_, err := version.NewVersion(opt.tfVersion)
	if err != nil {
		return "", err
	}

	return downloadWithVerification(opt.tfVersion, opt.installDir)
}

func Find(opts ...ExecPathFinder) (string, error) {
	var terraformPath string

	// go through the options in order
	// until a valid terraform executable is found
	for _, opt := range opts {
		p, err := opt.ExecPath()
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

	err := runTerraformVersion(terraformPath)
	if err != nil {
		return "", fmt.Errorf("executable found at path %s is not terraform: %s", terraformPath, err)
	}

	if terraformPath == "" {
		return "", fmt.Errorf("could not find terraform executable")
	}

	return terraformPath, nil
}

func downloadWithVerification(tfVersion string, installDir string) (string, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// setup: ensure we have a place to put our downloaded terraform binary
	var tfDir string
	var err error
	if installDir == "" {
		tfDir, err = ioutil.TempDir("", "tfexec")
		if err != nil {
			return "", fmt.Errorf("failed to create temp dir: %s", err)
		}
	} else {
		if _, err := os.Stat(installDir); err != nil {
			return "", fmt.Errorf("could not access directory %s for installing Terraform: %s", installDir, err)
		}
		tfDir = installDir

	}

	// setup: getter client
	httpHeader := make(http.Header)
	httpHeader.Set("User-Agent", "HashiCorp-tfinstall/"+Version)
	httpGetter := &getter.HttpGetter{
		Netrc: true,
	}
	client := getter.Client{
		Getters: map[string]getter.Getter{
			"https": httpGetter,
		},
	}
	client.Mode = getter.ClientModeAny

	// firstly, download and verify the signature of the checksum file

	sumsTmpDir, err := ioutil.TempDir("", "tfinstall")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(sumsTmpDir)

	sumsFilename := "terraform_" + tfVersion + "_SHA256SUMS"
	sumsSigFilename := sumsFilename + ".sig"

	sumsUrl := fmt.Sprintf("%s/%s/%s",
		baseUrl, tfVersion, sumsFilename)
	sumsSigUrl := fmt.Sprintf("%s/%s/%s",
		baseUrl, tfVersion, sumsSigFilename)

	client.Src = sumsUrl
	client.Dst = sumsTmpDir
	err = client.Get()
	if err != nil {
		return "", fmt.Errorf("error fetching checksums: %s", err)
	}

	client.Src = sumsSigUrl
	err = client.Get()
	if err != nil {
		return "", fmt.Errorf("error fetching checksums signature: %s", err)
	}

	sumsPath := filepath.Join(sumsTmpDir, sumsFilename)
	sumsSigPath := filepath.Join(sumsTmpDir, sumsSigFilename)

	err = verifySumsSignature(sumsPath, sumsSigPath)
	if err != nil {
		return "", err
	}

	// secondly, download Terraform itself, verifying the checksum
	url := tfUrl(tfVersion, osName, archName)
	client.Src = url
	client.Dst = tfDir
	client.Mode = getter.ClientModeDir
	err = client.Get()
	if err != nil {
		return "", err
	}

	return filepath.Join(tfDir, "terraform"), nil
}

func tfUrl(tfVersion, osName, archName string) string {
	sumsFilename := "terraform_" + tfVersion + "_SHA256SUMS"
	sumsUrl := fmt.Sprintf("%s/%s/%s",
		baseUrl, tfVersion, sumsFilename)
	return fmt.Sprintf(
		"%s/%s/terraform_%s_%s_%s.zip?checksum=file:%s",
		baseUrl, tfVersion, tfVersion, osName, archName, sumsUrl,
	)
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

// verifySumsSignature downloads SHA256SUMS and SHA256SUMS.sig and verifies
// the signature using the HashiCorp public key.
func verifySumsSignature(sumsPath, sumsSigPath string) error {
	el, err := openpgp.ReadArmoredKeyRing(strings.NewReader(hashicorpPublicKey))
	if err != nil {
		return err
	}
	data, err := os.Open(sumsPath)
	if err != nil {
		return err
	}
	sig, err := os.Open(sumsSigPath)
	if err != nil {
		return err
	}
	_, err = openpgp.CheckDetachedSignature(el, data, sig)

	return err
}

func runTerraformVersion(execPath string) error {
	cmd := exec.Command(execPath, "version")

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(string(out), "Terraform v") {
		return fmt.Errorf("located executable at %s, but output of `terraform version` was:\n%s", execPath, out)
	}

	return nil
}

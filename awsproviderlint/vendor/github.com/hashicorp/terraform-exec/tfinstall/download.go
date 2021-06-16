package tfinstall

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
	"golang.org/x/crypto/openpgp"
)

func ensureInstallDir(installDir string) (string, error) {
	if installDir == "" {
		return ioutil.TempDir("", "tfexec")
	}

	if _, err := os.Stat(installDir); err != nil {
		return "", fmt.Errorf("could not access directory %s for installing Terraform: %w", installDir, err)
	}

	return installDir, nil
}

func downloadWithVerification(ctx context.Context, tfVersion string, installDir string, appendUserAgent string) (string, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// setup: ensure we have a place to put our downloaded terraform binary
	tfDir, err := ensureInstallDir(installDir)
	if err != nil {
		return "", err
	}

	httpGetter := &getter.HttpGetter{
		Netrc:  true,
		Client: newHTTPClient(appendUserAgent),
	}
	client := getter.Client{
		Ctx: ctx,
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
	sumsSigFilename := sumsFilename + ".72D7468F.sig"

	sumsURL := fmt.Sprintf("%s/%s/%s", baseURL, tfVersion, sumsFilename)
	sumsSigURL := fmt.Sprintf("%s/%s/%s", baseURL, tfVersion, sumsSigFilename)

	client.Src = sumsURL
	client.Dst = sumsTmpDir
	err = client.Get()
	if err != nil {
		return "", fmt.Errorf("error fetching checksums at URL %s: %w", sumsURL, err)
	}

	client.Src = sumsSigURL
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
	url := tfURL(tfVersion, osName, archName)
	client.Src = url
	client.Dst = tfDir
	client.Mode = getter.ClientModeDir
	err = client.Get()
	if err != nil {
		return "", err
	}

	return filepath.Join(tfDir, "terraform"), nil
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

func tfURL(tfVersion, osName, archName string) string {
	sumsFilename := "terraform_" + tfVersion + "_SHA256SUMS"
	sumsURL := fmt.Sprintf("%s/%s/%s", baseURL, tfVersion, sumsFilename)
	return fmt.Sprintf(
		"%s/%s/terraform_%s_%s_%s.zip?checksum=file:%s",
		baseURL, tfVersion, tfVersion, osName, archName, sumsURL,
	)
}

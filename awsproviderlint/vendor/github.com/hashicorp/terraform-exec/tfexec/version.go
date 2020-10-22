package tfexec

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
)

var (
	tf0_12_0 = version.Must(version.NewVersion("0.12.0"))
	tf0_13_0 = version.Must(version.NewVersion("0.13.0"))
)

// Version returns structured output from the terraform version command including both the Terraform CLI version
// and any initialized provider versions. This will read cached values when present unless the skipCache parameter
// is set to true.
func (tf *Terraform) Version(ctx context.Context, skipCache bool) (tfVersion *version.Version, providerVersions map[string]*version.Version, err error) {
	tf.versionLock.Lock()
	defer tf.versionLock.Unlock()

	if tf.execVersion == nil || skipCache {
		tf.execVersion, tf.provVersions, err = tf.version(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	return tf.execVersion, tf.provVersions, nil
}

// version does not use the locking on the Terraform instance and should probably not be used directly, prefer Version.
func (tf *Terraform) version(ctx context.Context) (*version.Version, map[string]*version.Version, error) {
	// TODO: 0.13.0-beta2? and above supports a `-json` on the version command, should add support
	// for that here and fallback to string parsing

	versionCmd := tf.buildTerraformCmd(ctx, nil, "version")

	var outBuf bytes.Buffer
	versionCmd.Stdout = &outBuf

	err := tf.runTerraformCmd(versionCmd)
	if err != nil {
		return nil, nil, err
	}

	tfVersion, providerVersions, err := parseVersionOutput(outBuf.String())
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse version: %w", err)
	}

	return tfVersion, providerVersions, nil
}

var (
	simpleVersionRe = `v?(?P<version>[0-9]+(?:\.[0-9]+)*(?:-[A-Za-z0-9\.]+)?)`

	versionOutputRe         = regexp.MustCompile(`Terraform ` + simpleVersionRe)
	providerVersionOutputRe = regexp.MustCompile(`(\n\+ provider[\. ](?P<name>\S+) ` + simpleVersionRe + `)`)
)

func parseVersionOutput(stdout string) (*version.Version, map[string]*version.Version, error) {
	stdout = strings.TrimSpace(stdout)

	submatches := versionOutputRe.FindStringSubmatch(stdout)
	if len(submatches) != 2 {
		return nil, nil, fmt.Errorf("unexpected number of version matches %d for %s", len(submatches), stdout)
	}
	v, err := version.NewVersion(submatches[1])
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse version %q: %w", submatches[1], err)
	}

	allSubmatches := providerVersionOutputRe.FindAllStringSubmatch(stdout, -1)
	provV := map[string]*version.Version{}

	for _, submatches := range allSubmatches {
		if len(submatches) != 4 {
			return nil, nil, fmt.Errorf("unexpected number of provider version matches %d for %s", len(submatches), stdout)
		}

		v, err := version.NewVersion(submatches[3])
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse provider version %q: %w", submatches[3], err)
		}

		provV[submatches[2]] = v
	}

	return v, provV, err
}

func errorVersionString(v *version.Version) string {
	if v == nil {
		return "-"
	}
	return v.String()
}

// compatible asserts compatibility of the cached terraform version with the executable, and returns a well known error if not.
func (tf *Terraform) compatible(ctx context.Context, minInclusive *version.Version, maxExclusive *version.Version) error {
	tfv, _, err := tf.Version(ctx, false)
	if err != nil {
		return err
	}
	if ok := versionInRange(tfv, minInclusive, maxExclusive); !ok {
		return &ErrVersionMismatch{
			MinInclusive: errorVersionString(minInclusive),
			MaxExclusive: errorVersionString(maxExclusive),
			Actual:       errorVersionString(tfv),
		}
	}

	return nil
}

func stripPrereleaseAndMeta(v *version.Version) *version.Version {
	if v == nil {
		return nil
	}
	segs := []string{}
	for _, s := range v.Segments() {
		segs = append(segs, strconv.Itoa(s))
	}
	vs := strings.Join(segs, ".")
	clean, _ := version.NewVersion(vs)
	return clean
}

// versionInRange checks compatibility of the Terraform version. The minimum is inclusive and the max
// is exclusive, equivalent to min <= expected version < max.
//
// Pre-release information is ignored for comparison.
func versionInRange(tfv *version.Version, minInclusive *version.Version, maxExclusive *version.Version) bool {
	if minInclusive == nil && maxExclusive == nil {
		return true
	}
	tfv = stripPrereleaseAndMeta(tfv)
	minInclusive = stripPrereleaseAndMeta(minInclusive)
	maxExclusive = stripPrereleaseAndMeta(maxExclusive)
	if minInclusive != nil && !tfv.GreaterThanOrEqual(minInclusive) {
		return false
	}
	if maxExclusive != nil && !tfv.LessThan(maxExclusive) {
		return false
	}

	return true
}

// The meta package provides a location to set the release version
// and any other relevant metadata for the SDK.
//
// This package should not import any other SDK packages.
package meta

import (
	"fmt"

	version "github.com/hashicorp/go-version"
)

// The main version number that is being run at the moment.
var SDKVersion = "2.4.2"

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
var SDKPrerelease = ""

// SemVer is an instance of version.Version. This has the secondary
// benefit of verifying during tests and init time that our version is a
// proper semantic version, which should always be the case.
var SemVer *version.Version

func init() {
	SemVer = version.Must(version.NewVersion(SDKVersion))
}

// VersionString returns the complete version string, including prerelease
func SDKVersionString() string {
	if SDKPrerelease != "" {
		return fmt.Sprintf("%s-%s", SDKVersion, SDKPrerelease)
	}
	return SDKVersion
}

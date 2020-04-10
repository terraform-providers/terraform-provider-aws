package cmdflags

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bflad/tfproviderlint/version"
)

// AddVersionFlag adds -V and -version flags to commands
func AddVersionFlag() {
	flag.Var(versionFlag{}, "V", "print version and exit")
	flag.Var(versionFlag{}, "version", "print version and exit")
}

type versionFlag struct{}

func (versionFlag) IsBoolFlag() bool { return true }
func (versionFlag) Get() interface{} { return nil }
func (versionFlag) String() string   { return "" }
func (versionFlag) Set(s string) error {
	name := os.Args[0]
	name = name[strings.LastIndex(name, `/`)+1:]
	name = name[strings.LastIndex(name, `\`)+1:]
	name = strings.TrimSuffix(name, ".exe")

	// The go command uses -V=full to get a unique identifier for this tool.
	// Use a fully specified version in that case.
	fmt.Printf("%s %s\n", name, version.GetVersion().VersionNumber(s == "full"))
	os.Exit(0)

	return nil
}

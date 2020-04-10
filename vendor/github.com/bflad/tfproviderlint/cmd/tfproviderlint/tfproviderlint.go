// The tfproviderlint command is a static checker for Terraform Providers.
//
// Each analyzer flag name is preceded by the analyzer name: -NAME.flag.
// In addition, the -NAME flag itself controls whether the
// diagnostics of that analyzer are displayed. (A disabled analyzer may yet
// be run if it is required by some other analyzer that is enabled.)
package main

import (
	"github.com/bflad/tfproviderlint/helper/cmdflags"
	"github.com/bflad/tfproviderlint/passes"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	cmdflags.AddVersionFlag()

	multichecker.Main(passes.AllChecks...)
}

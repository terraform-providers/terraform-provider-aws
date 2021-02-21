package R019

import (
	"flag"
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatahaschangescallexpr"
)

const Doc = `check for (*schema.ResourceData).HasChanges() calls with many arguments

The R019 analyzer reports when there are a large number of arguments being
passed to (*schema.ResourceData).HasChanges(), which it may be preferable to
use (*schema.ResourceData).HasChangesExcept() instead.

Optional parameters:
  -threshold=5 Number of arguments for reporting
`

const analyzerName = "R019"

var (
	threshold int
)

var Analyzer = &analysis.Analyzer{
	Name:  analyzerName,
	Doc:   Doc,
	Flags: parseFlags(),
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		resourcedatahaschangescallexpr.Analyzer,
	},
	Run: run,
}

func parseFlags() flag.FlagSet {
	var flags = flag.NewFlagSet(analyzerName, flag.ExitOnError)
	flags.IntVar(&threshold, "threshold", 5, "Number of arguments for reporting")
	return *flags
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	callExprs := pass.ResultOf[resourcedatahaschangescallexpr.Analyzer].([]*ast.CallExpr)
	for _, callExpr := range callExprs {
		if ignorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		if len(callExpr.Args) < threshold {
			continue
		}

		pass.Reportf(callExpr.Pos(), "%s: d.HasChanges() has many arguments, consider d.HasChangesExcept()", analyzerName)
	}

	return nil, nil
}

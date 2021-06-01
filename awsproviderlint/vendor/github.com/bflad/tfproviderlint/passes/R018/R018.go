package R018

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/stdlib/timesleepcallexpr"
)

const Doc = `check for time.Sleep() function usage

Terraform Providers should generally avoid this function when waiting for API operations and prefer polling methods such as resource.Retry() or (resource.StateChangeConf).WaitForState().`

const analyzerName = "R018"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		timesleepcallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	callExprs := pass.ResultOf[timesleepcallexpr.Analyzer].([]*ast.CallExpr)
	for _, callExpr := range callExprs {
		if ignorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		pass.Reportf(callExpr.Pos(), "%s: prefer resource.Retry() or (resource.StateChangeConf).WaitForState() over time.Sleep()", analyzerName)
	}

	return nil, nil
}

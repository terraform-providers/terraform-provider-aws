// Package AT006 defines an Analyzer that checks for
// acceptance tests containing multiple resource.Test() invocations
package AT006

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/testaccfuncdecl"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for acceptance tests containing multiple resource.Test() invocations

The AT006 analyzer reports acceptance test functions that contain multiple
resource.Test() invocations. Acceptance tests should be split by invocation.`

const analyzerName = "AT006"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testaccfuncdecl.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	testFuncs := pass.ResultOf[testaccfuncdecl.Analyzer].([]*ast.FuncDecl)

	for _, testFunc := range testFuncs {
		if ignorer.ShouldIgnore(analyzerName, testFunc) {
			continue
		}

		var resourceTestInvocations int

		ast.Inspect(testFunc.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			if resource.IsFunc(callExpr.Fun, pass.TypesInfo, resource.FuncNameTest) {
				resourceTestInvocations += 1
			}

			if resourceTestInvocations > 1 {
				pass.Reportf(testFunc.Pos(), "%s: acceptance test function should contain only one Test invocation", analyzerName)
				return false
			}

			return true
		})

	}

	return nil, nil
}

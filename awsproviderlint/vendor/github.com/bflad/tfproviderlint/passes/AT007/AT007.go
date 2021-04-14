// Package AT007 defines an Analyzer that checks for
// acceptance tests containing multiple resource.ParallelTest() invocations
package AT007

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/testaccfuncdecl"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for acceptance tests containing multiple resource.ParallelTest() invocations

The AT007 analyzer reports acceptance test functions that contain multiple
resource.ParallelTest() invocations. Acceptance tests should be split by
invocation and multiple resource.ParallelTest() will cause a panic.`

const analyzerName = "AT007"

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

		var resourceParallelTestInvocations int

		ast.Inspect(testFunc.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			if resource.IsFunc(callExpr.Fun, pass.TypesInfo, resource.FuncNameParallelTest) {
				resourceParallelTestInvocations += 1
			}

			if resourceParallelTestInvocations > 1 {
				pass.Reportf(testFunc.Pos(), "%s: acceptance test function should contain only one ParallelTest invocation", analyzerName)
				return false
			}

			return true
		})

	}

	return nil, nil
}

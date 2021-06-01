// Package AT005 defines an Analyzer that checks for
// acceptance tests prefixed with Test but not TestAcc
package AT005

import (
	"go/ast"
	"strings"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/testfuncdecl"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for acceptance test function names missing TestAcc prefix

The AT005 analyzer reports test function names (Test prefix) that contain
resource.Test() or resource.ParallelTest(), which should be named with
the TestAcc prefix.`

const analyzerName = "AT005"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testfuncdecl.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	testFuncs := pass.ResultOf[testfuncdecl.Analyzer].([]*ast.FuncDecl)

	for _, testFunc := range testFuncs {
		if ignorer.ShouldIgnore(analyzerName, testFunc) {
			continue
		}

		if strings.HasPrefix(testFunc.Name.Name, "TestAcc") {
			continue
		}

		ast.Inspect(testFunc.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			isResourceTest := resource.IsFunc(callExpr.Fun, pass.TypesInfo, resource.FuncNameTest)
			isResourceParallelTest := resource.IsFunc(callExpr.Fun, pass.TypesInfo, resource.FuncNameParallelTest)

			if !isResourceTest && !isResourceParallelTest {
				return true
			}

			pass.Reportf(testFunc.Pos(), "%s: acceptance test function name should begin with TestAcc", analyzerName)
			return true
		})

	}

	return nil, nil
}

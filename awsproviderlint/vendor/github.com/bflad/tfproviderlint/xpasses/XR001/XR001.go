// Package XR001 defines an Analyzer that checks for
// ResourceData.Set() calls using * dereferences
package XR001

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatagetokexistscallexpr"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for ResourceData.GetOkExists() calls

The XR001 analyzer reports usage of GetOkExists() calls, which generally do
not work as expected. Usage should be moved to standard Get() and GetOk()
calls.`

const analyzerName = "XR001"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		resourcedatagetokexistscallexpr.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	callExprs := pass.ResultOf[resourcedatagetokexistscallexpr.Analyzer].([]*ast.CallExpr)
	for _, callExpr := range callExprs {
		if ignorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		pass.Reportf(callExpr.Pos(), "%s: ResourceData.GetOkExists() call should be avoided", analyzerName)
	}

	return nil, nil
}

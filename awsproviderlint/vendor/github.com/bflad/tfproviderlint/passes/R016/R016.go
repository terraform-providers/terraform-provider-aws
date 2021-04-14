package R016

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatasetidcallexpr"
)

const Doc = `check for (*schema.ResourceData).SetId() usage with unstable resource.PrefixedUniqueId() value

Schema attributes should be stable across Terraform runs.`

const analyzerName = "R016"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		resourcedatasetidcallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	callExprs := pass.ResultOf[resourcedatasetidcallexpr.Analyzer].([]*ast.CallExpr)
	for _, callExpr := range callExprs {
		if ignorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		if len(callExpr.Args) < 1 {
			continue
		}

		ast.Inspect(callExpr.Args[0], func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			if resource.IsFunc(callExpr.Fun, pass.TypesInfo, resource.FuncNamePrefixedUniqueId) {
				pass.Reportf(callExpr.Pos(), "%s: schema attributes should be stable across Terraform runs", analyzerName)
				return false
			}

			return true
		})
	}

	return nil, nil
}

// Package R001 defines an Analyzer that checks for
// ResourceData.Set() calls using complex key argument
package R001

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatasetcallexpr"
)

const Doc = `check for ResourceData.Set() calls using complex key argument

The R001 analyzer reports a complex key argument for a Set()
call. It is preferred to explicitly use a string literal as the key argument.`

const analyzerName = "R001"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		resourcedatasetcallexpr.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	sets := pass.ResultOf[resourcedatasetcallexpr.Analyzer].([]*ast.CallExpr)
	for _, set := range sets {
		if ignorer.ShouldIgnore(analyzerName, set) {
			continue
		}

		if len(set.Args) < 2 {
			continue
		}

		switch v := set.Args[0].(type) {
		default:
			pass.Reportf(v.Pos(), "%s: ResourceData.Set() key argument should be string literal", analyzerName)
		case *ast.BasicLit:
			continue
		}
	}

	return nil, nil
}

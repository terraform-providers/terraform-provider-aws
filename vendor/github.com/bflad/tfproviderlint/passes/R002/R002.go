// Package R002 defines an Analyzer that checks for
// ResourceData.Set() calls using * dereferences
package R002

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatasetcallexpr"
)

const Doc = `check for ResourceData.Set() calls using * dereferences

The R002 analyzer reports likely extraneous uses of
star (*) dereferences for a Set() call. The Set() function automatically
handles pointers and * dereferences without nil checks can panic.`

const analyzerName = "R002"

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

		switch v := set.Args[1].(type) {
		default:
			continue
		case *ast.StarExpr:
			pass.Reportf(v.Pos(), "%s: ResourceData.Set() pointer value dereference is extraneous", analyzerName)
		}
	}

	return nil, nil
}

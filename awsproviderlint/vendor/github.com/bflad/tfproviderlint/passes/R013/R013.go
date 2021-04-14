package R013

import (
	"go/ast"
	"strings"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcemapcompositelit"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for resource names that do not contain at least one underscore

The R013 analyzer reports cases of resource names which do not include at least
one underscore character (_). Resources should be named with the provider name
and API resource name separated by an underscore to clarify where a resource is
declared and configured.`

const analyzerName = "R013"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		resourcemapcompositelit.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	compositeLits := pass.ResultOf[resourcemapcompositelit.Analyzer].([]*ast.CompositeLit)

	for _, compositeLit := range compositeLits {
		if ignorer.ShouldIgnore(analyzerName, compositeLit) {
			continue
		}

		for _, expr := range schema.GetResourceMapResourceNames(compositeLit) {
			resourceName := astutils.ExprStringValue(expr)

			if resourceName == nil {
				continue
			}

			if strings.ContainsAny(*resourceName, "_") {
				continue
			}

			pass.Reportf(expr.Pos(), "%s: resource names should include the provider name and at least one underscore (_)", analyzerName)
		}
	}

	return nil, nil
}

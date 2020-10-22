package XS002

import (
	"go/ast"
	"sort"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemamapcompositelit"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for Schema that attribute names are in alphabetical order

The XS002 analyzer reports cases of schemas where attribute names
are not in alphabetical order.`

const analyzerName = "XS002"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		schemamapcompositelit.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemamapcompositelits := pass.ResultOf[schemamapcompositelit.Analyzer].([]*ast.CompositeLit)

	for _, smap := range schemamapcompositelits {
		if ignorer.ShouldIgnore(analyzerName, smap) {
			continue
		}

		schemaKeys := make([]string, 0 , len(schema.GetSchemaMapAttributeNames(smap)))
		for _, attributeName := range schema.GetSchemaMapAttributeNames(smap) {
			if v := astutils.ExprStringValue(attributeName); v != nil {
				schemaKeys = append(schemaKeys, *v)
			}
		}

		if !sort.StringsAreSorted(schemaKeys) {
			pass.Reportf(smap.Pos(), "%s: schema attributes should be in alphabetical order", analyzerName)
		}
	}

	return nil, nil
}

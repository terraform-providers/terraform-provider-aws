// Package S015 defines an Analyzer that checks for
// Schema that attribute names contain only lowercase
// alphanumerics and underscores
package S015

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemamapcompositelit"
)

const Doc = `check for Schema that attribute names are valid

The S015 analyzer reports cases of schemas which the attribute name
includes characters outside lowercase alphanumerics and underscores,
which will fail provider schema validation.`

const analyzerName = "S015"

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

		for _, attributeName := range schema.GetSchemaMapAttributeNames(smap) {
			switch t := attributeName.(type) {
			default:
				continue
			case *ast.BasicLit:
				value := strings.Trim(t.Value, `"`)

				if !schema.AttributeNameRegexp.MatchString(value) {
					pass.Reportf(t.Pos(), "%s: schema attribute names should only be lowercase alphanumeric characters or underscores", analyzerName)
				}
			}
		}
	}

	return nil, nil
}

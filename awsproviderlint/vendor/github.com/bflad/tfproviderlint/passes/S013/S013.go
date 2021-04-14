// Package S013 defines an Analyzer that checks for
// Schema that one of Computed, Optional, or Required
// is not configured
package S013

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemamapcompositelit"
)

const Doc = `check for Schema that are missing required fields

The S013 analyzer reports cases of schemas which one of Computed,
Optional, or Required is not configured, which will fail provider
schema validation.`

const analyzerName = "S013"

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
		for _, schemaCompositeLit := range schema.GetSchemaMapSchemas(smap) {
			schemaInfo := schema.NewSchemaInfo(schemaCompositeLit, pass.TypesInfo)

			if ignorer.ShouldIgnore(analyzerName, schemaInfo.AstCompositeLit) {
				continue
			}

			if schemaInfo.Schema.Computed || schemaInfo.Schema.Optional || schemaInfo.Schema.Required {
				continue
			}

			switch t := schemaInfo.AstCompositeLit.Type.(type) {
			default:
				pass.Reportf(schemaInfo.AstCompositeLit.Lbrace, "%s: schema should configure one of Computed, Optional, or Required", analyzerName)
			case *ast.SelectorExpr:
				pass.Reportf(t.Sel.Pos(), "%s: schema should configure one of Computed, Optional, or Required", analyzerName)
			}
		}
	}

	return nil, nil
}

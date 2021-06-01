// Package S007 defines an Analyzer that checks for
// Schema with Required enabled and ConflictsWith configured
package S007

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfo"
)

const Doc = `check for Schema with Required enabled and ConflictsWith configured

The S007 analyzer reports cases of schemas which enables Required
and configures ConflictsWith, which will fail provider schema validation.`

const analyzerName = "S007"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		schemainfo.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemaInfos := pass.ResultOf[schemainfo.Analyzer].([]*schema.SchemaInfo)
	for _, schemaInfo := range schemaInfos {
		if ignorer.ShouldIgnore(analyzerName, schemaInfo.AstCompositeLit) {
			continue
		}

		if !schemaInfo.Schema.Required || schemaInfo.Schema.ConflictsWith == nil {
			continue
		}

		switch t := schemaInfo.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schemaInfo.AstCompositeLit.Lbrace, "%s: schema should not enable Required and configure ConflictsWith", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema should not enable Required and configure ConflictsWith", analyzerName)
		}
	}

	return nil, nil
}

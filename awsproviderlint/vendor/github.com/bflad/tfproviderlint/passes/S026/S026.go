package S026

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfocomputedonly"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for Schema with only Computed enabled and ConflictsWith configured

The S026 analyzer reports cases of schemas which only enables Computed
and configures ConflictsWith, which is not valid.`

const analyzerName = "S026"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		schemainfocomputedonly.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemaInfos := pass.ResultOf[schemainfocomputedonly.Analyzer].([]*schema.SchemaInfo)
	for _, schemaInfo := range schemaInfos {
		if ignorer.ShouldIgnore(analyzerName, schemaInfo.AstCompositeLit) {
			continue
		}

		if schemaInfo.Schema.ConflictsWith == nil {
			continue
		}

		switch t := schemaInfo.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schemaInfo.AstCompositeLit.Lbrace, "%s: schema should not only enable Computed and configure ConflictsWith", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema should not only enable Computed and configure ConflictsWith", analyzerName)
		}
	}

	return nil, nil
}

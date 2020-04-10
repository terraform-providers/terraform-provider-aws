// Package S009 defines an Analyzer that checks for
// Schema of TypeList or TypeSet with ValidateFunc configured
package S009

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfo"
)

const Doc = `check for Schema of TypeList or TypeSet with ValidateFunc configured

The S009 analyzer reports cases of TypeList or TypeSet schemas with ValidateFunc configured,
which will fail schema validation.`

const analyzerName = "S009"

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

		if !schemaInfo.DeclaresField(schema.SchemaFieldValidateFunc) {
			continue
		}

		if !schemaInfo.IsOneOfTypes(schema.SchemaValueTypeList, schema.SchemaValueTypeSet) {
			continue
		}

		switch t := schemaInfo.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schemaInfo.AstCompositeLit.Lbrace, "%s: schema of TypeList or TypeSet should not include top level ValidateFunc", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema of TypeList or TypeSet should not include top level ValidateFunc", analyzerName)
		}
	}

	return nil, nil
}

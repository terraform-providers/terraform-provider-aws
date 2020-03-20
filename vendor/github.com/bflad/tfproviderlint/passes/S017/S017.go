// Package S017 defines an Analyzer that checks for
// Schema including MaxItems or MinItems without TypeList,
// TypeMap, or TypeSet
package S017

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfo"
)

const Doc = `check for Schema including MaxItems or MinItems without proper Type

The S017 analyzer reports cases of schema including MaxItems or MinItems without
TypeList, TypeMap, or TypeSet, which will fail schema validation.`

const analyzerName = "S017"

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

		if !schemaInfo.DeclaresField(schema.SchemaFieldMaxItems) && !schemaInfo.DeclaresField(schema.SchemaFieldMinItems) {
			continue
		}

		if schemaInfo.IsOneOfTypes(schema.SchemaValueTypeList, schema.SchemaValueTypeMap, schema.SchemaValueTypeSet) {
			continue
		}

		switch t := schemaInfo.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(schemaInfo.AstCompositeLit.Lbrace, "%s: schema MaxItems or MinItems should only be included for TypeList, TypeMap, or TypeSet", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema MaxItems or MinItems should only be included for TypeList, TypeMap, or TypeSet", analyzerName)
		}
	}

	return nil, nil
}

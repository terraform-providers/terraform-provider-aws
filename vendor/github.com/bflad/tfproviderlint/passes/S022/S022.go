// Package S022 defines an Analyzer that checks for
// Schema of TypeMap with invalid Elem of *schema.Resource
package S022

import (
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for Schema of TypeMap with invalid Elem of *schema.Resource

The S022 analyzer reports cases of schema that declare Elem of *schema.Resource
with TypeMap, which has undefined behavior. Only TypeList and TypeSet can be
used for configuration block attributes.`

const analyzerName = "S022"

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

		if !schemaInfo.IsType(schema.SchemaValueTypeMap) {
			continue
		}

		if !schemaInfo.DeclaresField(schema.SchemaFieldElem) {
			continue
		}

		elem := schemaInfo.Fields[schema.SchemaFieldElem].Value

		if !schema.IsTypeResource(pass.TypesInfo.TypeOf(elem)) {
			continue
		}

		pass.Reportf(elem.Pos(), "%s: schema of TypeMap should not use Elem of *schema.Resource", analyzerName)
	}

	return nil, nil
}

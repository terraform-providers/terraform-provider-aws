package S024

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourceinfodatasourceonly"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for Schema that should omit ForceNew in data source schema attributes

The S024 analyzer reports usage of ForceNew in data source schema attributes,
which is unnecessary.`

const analyzerName = "S024"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		resourceinfodatasourceonly.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	resourceInfos := pass.ResultOf[resourceinfodatasourceonly.Analyzer].([]*schema.ResourceInfo)
	for _, resourceInfo := range resourceInfos {
		if ignorer.ShouldIgnore(analyzerName, resourceInfo.AstCompositeLit) {
			continue
		}

		var schemaInfos []*schema.SchemaInfo

		ast.Inspect(resourceInfo.AstCompositeLit, func(n ast.Node) bool {
			compositeLit, ok := n.(*ast.CompositeLit)

			if !ok {
				return true
			}

			if schema.IsMapStringSchema(compositeLit, pass.TypesInfo) {
				for _, mapSchema := range schema.GetSchemaMapSchemas(compositeLit) {
					schemaInfos = append(schemaInfos, schema.NewSchemaInfo(mapSchema, pass.TypesInfo))
				}
			} else if schema.IsTypeSchema(pass.TypesInfo.TypeOf(compositeLit.Type)) {
				schemaInfos = append(schemaInfos, schema.NewSchemaInfo(compositeLit, pass.TypesInfo))
			}

			return true
		})

		for _, schemaInfo := range schemaInfos {
			if !schemaInfo.DeclaresField(schema.SchemaFieldForceNew) {
				continue
			}

			pass.Reportf(schemaInfo.Fields[schema.SchemaFieldForceNew].Pos(), "%s: ForceNew is extraneous in data source schema attributes", analyzerName)
		}
	}

	return nil, nil
}

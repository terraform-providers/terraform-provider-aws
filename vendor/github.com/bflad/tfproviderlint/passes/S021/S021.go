// Package S021 defines an Analyzer that checks for
// Schema that should omit ComputedWhen
package S021

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfo"
)

const Doc = `check for Schema that should omit ComputedWhen

The S021 analyzer reports cases of schema that declare ComputedWhen that should
be removed.`

const analyzerName = "S021"

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

		field := schema.SchemaFieldComputedWhen

		if schemaInfo.DeclaresField(field) {
			pass.Reportf(schemaInfo.Fields[field].Value.Pos(), "%s: schema should omit ComputedWhen", analyzerName)
		}
	}

	return nil, nil
}

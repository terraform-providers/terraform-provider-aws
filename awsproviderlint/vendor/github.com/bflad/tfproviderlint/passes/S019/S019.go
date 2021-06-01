// Package S019 defines an Analyzer that checks for
// Schema that should omit Computed, Optional, or Required
// set to false
package S019

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfo"
)

const Doc = `check for Schema that should omit Computed, Optional, or Required set to false

The S019 analyzer reports cases of schema that use Computed: false, Optional: false, or
Required: false that should be removed.`

const analyzerName = "S019"

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

		for _, field := range []string{schema.SchemaFieldComputed, schema.SchemaFieldOptional, schema.SchemaFieldRequired} {
			if schemaInfo.DeclaresBoolFieldWithZeroValue(field) {
				pass.Reportf(schemaInfo.Fields[field].Value.Pos(), "%s: schema should omit Computed, Optional, or Required set to false", analyzerName)
			}
		}
	}

	return nil, nil
}

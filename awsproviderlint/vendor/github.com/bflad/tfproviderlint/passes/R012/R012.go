package R012

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourceinfodatasourceonly"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for data source Resource with CustomizeDiff configured

The R012 analyzer reports cases of data sources which configure CustomizeDiff,
which is not valid.`

const analyzerName = "R012"

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

		if !resourceInfo.DeclaresField(schema.ResourceFieldCustomizeDiff) {
			continue
		}

		switch t := resourceInfo.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(resourceInfo.AstCompositeLit.Lbrace, "%s: data source should not configure CustomizeDiff", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: data source should not configure CustomizeDiff", analyzerName)
		}
	}

	return nil, nil
}

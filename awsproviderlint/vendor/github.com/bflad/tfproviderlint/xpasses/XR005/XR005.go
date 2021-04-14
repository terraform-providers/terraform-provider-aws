package XR005

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourceinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for Resource that Description is configured

The XR005 analyzer reports cases of resources where Description is not
configured, which is generally useful for providers that wish to
automatically generate documentation based on the schema information.`

const analyzerName = "XR005"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		resourceinfo.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	resourceInfos := pass.ResultOf[resourceinfo.Analyzer].([]*schema.ResourceInfo)
	for _, resourceInfo := range resourceInfos {
		if ignorer.ShouldIgnore(analyzerName, resourceInfo.AstCompositeLit) {
			continue
		}

		// Skip configuration block Elem
		if !resourceInfo.IsDataSource() && !resourceInfo.IsResource() {
			continue
		}

		if resourceInfo.Resource.Description != "" {
			continue
		}

		switch t := resourceInfo.AstCompositeLit.Type.(type) {
		default:
			pass.Reportf(resourceInfo.AstCompositeLit.Lbrace, "%s: resource should configure Description", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: resource should configure Description", analyzerName)
		}
	}

	return nil, nil
}

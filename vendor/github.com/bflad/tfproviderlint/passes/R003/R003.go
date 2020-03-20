// Package R003 defines an Analyzer that checks for
// Resource having Exists functions
package R003

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourceinfo"
)

const Doc = `check for Resource having Exists functions

The R003 analyzer reports likely extraneous uses of Exists
functions for a resource. Exists logic can be handled inside the Read function
to prevent logic duplication.`

const analyzerName = "R003"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		resourceinfo.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	resources := pass.ResultOf[resourceinfo.Analyzer].([]*schema.ResourceInfo)
	for _, resource := range resources {
		if ignorer.ShouldIgnore(analyzerName, resource.AstCompositeLit) {
			continue
		}

		kvExpr := resource.Fields[schema.ResourceFieldExists]

		if kvExpr == nil {
			continue
		}

		pass.Reportf(kvExpr.Key.Pos(), "%s: resource should not include Exists function", analyzerName)
	}

	return nil, nil
}

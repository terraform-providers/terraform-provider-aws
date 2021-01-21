package XR002

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourceinforesourceonly"
)

const Doc = `check for Resource that should implement Importer

The XR002 analyzer reports missing usage of Importer in resources.`

const analyzerName = "XR002"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		resourceinforesourceonly.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	resources := pass.ResultOf[resourceinforesourceonly.Analyzer].([]*schema.ResourceInfo)
	for _, resource := range resources {
		if ignorer.ShouldIgnore(analyzerName, resource.AstCompositeLit) {
			continue
		}

		if resource.DeclaresField(schema.ResourceFieldImporter) {
			continue
		}

		pass.Reportf(resource.AstCompositeLit.Pos(), "%s: resource should include Importer implementation", analyzerName)
	}

	return nil, nil
}

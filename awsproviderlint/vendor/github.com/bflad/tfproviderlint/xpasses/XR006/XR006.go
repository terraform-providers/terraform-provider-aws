package XR006

import (
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourceinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for Resource that implements Timeouts for missing Create, Delete, Read, or Update implementation

The XR006 analyzer reports extraneous Timeouts fields in resources where the
corresponding Create, Delete, Read, or Update implementation does not exist.`

const analyzerName = "XR006"

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
	resources := pass.ResultOf[resourceinfo.Analyzer].([]*schema.ResourceInfo)
	for _, resource := range resources {
		if ignorer.ShouldIgnore(analyzerName, resource.AstCompositeLit) {
			continue
		}

		if !resource.DeclaresField(schema.ResourceFieldCreate) && !resource.DeclaresField(schema.ResourceFieldCreateContext) && !resource.DeclaresField(schema.ResourceFieldCreateWithoutTimeout) && resource.Resource.Timeouts.Create != nil {
			pass.Reportf(resource.AstCompositeLit.Pos(), "%s: resource should not configure Timeouts.Create without Create implementation", analyzerName)
		}

		if !resource.DeclaresField(schema.ResourceFieldDelete) && !resource.DeclaresField(schema.ResourceFieldDeleteContext) && !resource.DeclaresField(schema.ResourceFieldDeleteWithoutTimeout) && resource.Resource.Timeouts.Delete != nil {
			pass.Reportf(resource.AstCompositeLit.Pos(), "%s: resource should not configure Timeouts.Delete without Delete implementation", analyzerName)
		}

		if !resource.DeclaresField(schema.ResourceFieldRead) && !resource.DeclaresField(schema.ResourceFieldReadContext) && !resource.DeclaresField(schema.ResourceFieldReadWithoutTimeout) && resource.Resource.Timeouts.Read != nil {
			pass.Reportf(resource.AstCompositeLit.Pos(), "%s: resource should not configure Timeouts.Read without Read implementation", analyzerName)
		}

		if !resource.DeclaresField(schema.ResourceFieldUpdate) && !resource.DeclaresField(schema.ResourceFieldUpdateContext) && !resource.DeclaresField(schema.ResourceFieldUpdateWithoutTimeout) && resource.Resource.Timeouts.Update != nil {
			pass.Reportf(resource.AstCompositeLit.Pos(), "%s: resource should not configure Timeouts.Update without Update implementation", analyzerName)
		}
	}

	return nil, nil
}

package resourceinfodatasourceonly

import (
	"reflect"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourceinfo"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "resourceinfodatasourceonly",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/schema.Resource literals of Data Sources for later passes",
	Requires: []*analysis.Analyzer{
		resourceinfo.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*schema.ResourceInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	resourceInfos := pass.ResultOf[resourceinfo.Analyzer].([]*schema.ResourceInfo)

	var result []*schema.ResourceInfo

	for _, resourceInfo := range resourceInfos {
		if !resourceInfo.IsDataSource() {
			continue
		}

		result = append(result, resourceInfo)
	}

	return result, nil
}

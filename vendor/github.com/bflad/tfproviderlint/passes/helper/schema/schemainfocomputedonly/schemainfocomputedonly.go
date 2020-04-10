package schemainfocomputedonly

import (
	"reflect"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfo"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "schemainfocomputedonly",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/schema.Schema literals with Computed: true only for later passes",
	Requires: []*analysis.Analyzer{
		schemainfo.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*schema.SchemaInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	schemaInfos := pass.ResultOf[schemainfo.Analyzer].([]*schema.SchemaInfo)

	var result []*schema.SchemaInfo

	for _, schemaInfo := range schemaInfos {
		if !schemaInfo.Schema.Computed || schemaInfo.Schema.Optional || schemaInfo.Schema.Required {
			continue
		}

		result = append(result, schemaInfo)
	}

	return result, nil
}

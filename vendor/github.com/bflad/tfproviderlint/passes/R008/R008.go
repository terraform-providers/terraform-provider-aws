package R008

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatasetpartialselectorexpr"
)

var Analyzer = analysisutils.DeprecatedReceiverMethodSelectorExprAnalyzer(
	"R008",
	resourcedatasetpartialselectorexpr.Analyzer,
	schema.PackageName,
	schema.TypeNameResourceData,
	"SetPartial",
)

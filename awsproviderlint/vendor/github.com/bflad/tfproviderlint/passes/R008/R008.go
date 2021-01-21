package R008

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatasetpartialcallexpr"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatasetpartialselectorexpr"
)

var Analyzer = analysisutils.DeprecatedReceiverMethodSelectorExprAnalyzer(
	"R008",
	resourcedatasetpartialcallexpr.Analyzer,
	resourcedatasetpartialselectorexpr.Analyzer,
	schema.PackagePath,
	schema.TypeNameResourceData,
	"SetPartial",
)

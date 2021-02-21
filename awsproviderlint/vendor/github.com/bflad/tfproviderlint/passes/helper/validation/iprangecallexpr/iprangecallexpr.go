package iprangecallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"iprangecallexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameIPRange,
)

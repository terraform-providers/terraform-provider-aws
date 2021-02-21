package stringinslicecallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"stringinslicecallexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameStringInSlice,
)

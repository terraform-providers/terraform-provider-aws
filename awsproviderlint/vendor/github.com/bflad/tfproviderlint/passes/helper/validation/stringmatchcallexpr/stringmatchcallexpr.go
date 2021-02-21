package stringmatchcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"stringmatchcallexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameStringMatch,
)

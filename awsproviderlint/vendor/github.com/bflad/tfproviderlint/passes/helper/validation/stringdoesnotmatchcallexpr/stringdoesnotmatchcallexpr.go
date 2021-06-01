package stringdoesnotmatchcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"stringdoesnotmatchcallexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameStringDoesNotMatch,
)

package V005

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/validatejsonstringselectorexpr"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V005",
	validatejsonstringselectorexpr.Analyzer,
	validation.PackagePath,
	validation.FuncNameValidateJsonString,
	validation.PackagePath,
	validation.FuncNameStringIsJSON,
)

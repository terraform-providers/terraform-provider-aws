package V007

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/validateregexpselectorexpr"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V007",
	validateregexpselectorexpr.Analyzer,
	validation.PackagePath,
	validation.FuncNameValidateRegexp,
	validation.PackagePath,
	validation.FuncNameStringIsValidRegExp,
)

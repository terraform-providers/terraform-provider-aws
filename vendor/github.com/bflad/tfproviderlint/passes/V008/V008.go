package V008

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/validaterfc3339timestringselectorexpr"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V008",
	validaterfc3339timestringselectorexpr.Analyzer,
	validation.PackagePath,
	validation.FuncNameValidateRFC3339TimeString,
	validation.PackagePath,
	validation.FuncNameIsRFC3339Time,
)

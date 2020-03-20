package V004

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/singleipselectorexpr"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V004",
	singleipselectorexpr.Analyzer,
	validation.PackageName,
	validation.FuncNameSingleIP,
	validation.PackageName,
	validation.FuncNameIsIPAddress,
)

package V003

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/iprangeselectorexpr"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V003",
	iprangeselectorexpr.Analyzer,
	validation.PackageName,
	validation.FuncNameIPRange,
	validation.PackageName,
	validation.FuncNameIsIPv4Range,
)

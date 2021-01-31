package V002

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/cidrnetworkselectorexpr"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V002",
	cidrnetworkselectorexpr.Analyzer,
	validation.PackagePath,
	validation.FuncNameCIDRNetwork,
	validation.PackagePath,
	validation.FuncNameIsCIDRNetwork,
)

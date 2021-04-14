package cidrnetworkselectorexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.SelectorExprAnalyzer(
	"cidrnetworkselectorexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameCIDRNetwork,
)

package V006

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
	"github.com/bflad/tfproviderlint/passes/helper/validation/validatelistuniquestringsselectorexpr"
)

var Analyzer = analysisutils.DeprecatedWithReplacementSelectorExprAnalyzer(
	"V006",
	validatelistuniquestringsselectorexpr.Analyzer,
	validation.PackageName,
	validation.FuncNameValidateListUniqueStrings,
	validation.PackageName,
	validation.FuncNameListOfUniqueStrings,
)

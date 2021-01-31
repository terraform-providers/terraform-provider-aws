package validatelistuniquestringsselectorexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/validation"
)

var Analyzer = analysisutils.SelectorExprAnalyzer(
	"validatelistuniquestringsselectorexpr",
	validation.IsFunc,
	validation.PackagePath,
	validation.FuncNameValidateListUniqueStrings,
)

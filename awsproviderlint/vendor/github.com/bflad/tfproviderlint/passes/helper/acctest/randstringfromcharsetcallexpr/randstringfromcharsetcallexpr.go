package randstringfromcharsetcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/acctest"
)

var Analyzer = analysisutils.FunctionCallExprAnalyzer(
	"randstringfromcharsetcallexpr",
	acctest.IsFunc,
	acctest.PackagePath,
	acctest.FuncNameRandStringFromCharSet,
)

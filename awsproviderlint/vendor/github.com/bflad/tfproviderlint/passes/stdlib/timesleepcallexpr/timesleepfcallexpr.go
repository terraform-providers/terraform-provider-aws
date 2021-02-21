package timesleepcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"timesleepcallexpr",
	"time",
	"Sleep",
)

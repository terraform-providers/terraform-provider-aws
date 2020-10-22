package fmtsprintfcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"fmtsprintfcallexpr",
	"fmt",
	"Sprintf",
)

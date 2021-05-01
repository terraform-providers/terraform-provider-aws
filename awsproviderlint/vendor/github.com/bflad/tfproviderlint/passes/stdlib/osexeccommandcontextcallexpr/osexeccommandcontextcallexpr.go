package osexeccommandcontextcallexpr

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
)

var Analyzer = analysisutils.StdlibFunctionCallExprAnalyzer(
	"osexeccommandcontextcallexpr",
	"os/exec",
	"CommandContext",
)

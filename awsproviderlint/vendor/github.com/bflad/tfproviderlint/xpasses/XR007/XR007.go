package XR007

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/passes/stdlib/osexeccommandcallexpr"
	"github.com/bflad/tfproviderlint/passes/stdlib/osexeccommandselectorexpr"
)

var Analyzer = analysisutils.AvoidSelectorExprAnalyzer(
	"XR007",
	osexeccommandcallexpr.Analyzer,
	osexeccommandselectorexpr.Analyzer,
	"os/exec",
	"Command",
)

package XR008

import (
	"github.com/bflad/tfproviderlint/helper/analysisutils"
	"github.com/bflad/tfproviderlint/passes/stdlib/osexeccommandcontextcallexpr"
	"github.com/bflad/tfproviderlint/passes/stdlib/osexeccommandcontextselectorexpr"
)

var Analyzer = analysisutils.AvoidSelectorExprAnalyzer(
	"XR008",
	osexeccommandcontextcallexpr.Analyzer,
	osexeccommandcontextselectorexpr.Analyzer,
	"os/exec",
	"CommandContext",
)

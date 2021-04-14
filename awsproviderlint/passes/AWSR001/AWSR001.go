package AWSR001

import (
	"go/ast"
	"strings"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/stdlib/fmtsprintfcallexpr"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for fmt.Sprintf() using amazonaws.com domain suffix

The AWSR001 analyzer reports when a fmt.Sprintf() call contains the
ending string ".amazonaws.com". This domain suffix is only valid in
the AWS Commercial and GovCloud (US) partitions.

To ensure the correct domain suffix is used in all partitions, the
AWSClient available to all resources provides the PartitionHostname()
and RegionalHostname() receiver methods.
`

const analyzerName = "AWSR001"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		fmtsprintfcallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	callExprs := pass.ResultOf[fmtsprintfcallexpr.Analyzer].([]*ast.CallExpr)
	commentIgnorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)

	for _, callExpr := range callExprs {
		if commentIgnorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		formatString := astutils.ExprStringValue(callExpr.Args[0])

		if formatString == nil {
			continue
		}

		if !strings.HasSuffix(*formatString, ".amazonaws.com") {
			continue
		}

		pass.Reportf(callExpr.Pos(), "%s: prefer (*AWSClient).PartitionHostname() or (*AWSClient).RegionalHostname()", analyzerName)
	}

	return nil, nil
}
